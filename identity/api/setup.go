package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	goth_github "github.com/markbates/goth/providers/github"
	"github.com/netlify/git-gateway/identity/models"
	"github.com/netlify/git-gateway/identity/secrets"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const setupTemplate = "setup.html"

func (a *API) setup(w http.ResponseWriter, r *http.Request) error {
	// GitHub App installation is complete and redirects back to us
	q := r.URL.Query()
	if q.Has("code") {
		client := github.NewClient(nil)
		config, _, err := client.Apps.CompleteAppManifest(context.TODO(), q.Get("code"))
		if err != nil {
			return err
		}
		app := &models.App{
			ID:            config.GetID(),
			Name:          config.GetName(),
			PEM:           []byte(config.GetPEM()),
			WebhookSecret: config.GetWebhookSecret(),
			Owner: models.AppOwner{
				Login: config.GetOwner().GetLogin(),
				Type:  config.GetOwner().GetType(),
			},
			ClientID:     config.GetClientID(),
			ClientSecret: config.GetClientSecret(),
		}

		if a.config.MultiInstanceMode {
			// MultiInstanceMode saves app in the database
			err = a.db.CreateApp(app)
			if err != nil {
				return err
			}
		} else {
			// SingleInstanceMode saves app in a SecretManager Secret
			err := secrets.SetApp(context.TODO(), os.Getenv("GITGATEWAY_GCP_SECRET"), app)
			if err != nil {
				return err
			}
			a.GetSingleApp = func() (*models.App, error) { return app, nil }
		}

		// Forward the user to install the new app
		a.installApp(w, r, app)
		return nil
	}

	// Start the setup flow
	return withTemplate(w, setupTemplate, func(t *template.Template) interface{} {
		data := make(map[string]interface{})
		data["State"] = "todo-manifest"
		data["App"] = map[string]interface{}{
			"Url":         withScheme(a.config.API.Endpoint),
			"HookUrl":     withScheme(singleJoiningSlash(a.config.API.Endpoint, AppHook)),
			"RedirectUrl": withScheme(singleJoiningSlash(a.config.API.Endpoint, AppSetupRedirectPath)),
			"CallbackUrl": withScheme(singleJoiningSlash(a.config.API.Endpoint, AppOAuthCallback)),
		}
		return data
	})
}

func (a *API) installApp(w http.ResponseWriter, r *http.Request, app *models.App) {
	state := url.Values{"app": []string{strconv.FormatInt(app.ID, 10)}}.Encode()
	redirectURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s", url.PathEscape(app.Name), url.QueryEscape(state))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (a *API) hook(w http.ResponseWriter, r *http.Request) error {
	// TODO validate webhook signature
	defer r.Body.Close()

	switch r.Header.Get("X-GitHub-Event") {
	case "installation":
		event := github.InstallationEvent{}
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			return err
		}
		logrus.Infof("Installation created for app %s (%s) with id %s", event.Installation.GetAppID(), event.Installation.GetAppSlug(), event.Installation.GetID())
		// err = a.db.CreateInstallation(&models.Installation{
		// 	AppID:          event.Installation.GetAppID(),
		// 	InstallationID: event.Installation.GetID(),
		// })
	}

	return nil
}

func (a *API) appClient(app *models.App) (*github.Client, error) {
	tr := http.DefaultTransport
	appsTr, err := ghinstallation.NewAppsTransport(tr, app.ID, app.PEM)
	client := github.NewClient(&http.Client{Transport: appsTr})
	return client, err
}

func (a *API) installationClient(app *models.App, installation *models.Installation) (*github.Client, error) {
	tr := http.DefaultTransport
	appsTr, err := ghinstallation.NewAppsTransport(tr, app.ID, app.PEM)
	installationTr := ghinstallation.NewFromAppsTransport(appsTr, installation.InstallationID)
	client := github.NewClient(&http.Client{Transport: installationTr})
	return client, err
}

func (a *API) githubConfig(app *models.App) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  goth_github.AuthURL,
			TokenURL: goth_github.TokenURL,
		},
		RedirectURL: withScheme(singleJoiningSlash(a.config.API.Endpoint, AppOAuthCallback)),
		Scopes:      nil,
	}
}

func (a *API) getApp(hint getAppHint) (*models.App, error) {
	if a.config.MultiInstanceMode {
		id, err := hint.getAppID()
		if err != nil {
			return nil, err
		}
		return a.db.GetApp(id)
	} else {
		return a.GetSingleApp()
	}
}

type getAppHint interface {
	getAppID() (int64, error)
}

type getAppIdFromState struct {
	*http.Request
}

func (hint getAppIdFromState) getAppID() (id int64, err error) {
	// Get the app from the url state context
	state := hint.URL.Query().Get("state")
	stateQuery, err := url.ParseQuery(state)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(stateQuery.Get("app"), 10, 64)
}

func (a *API) callback(w http.ResponseWriter, r *http.Request) error {
	app, err := a.getApp(getAppIdFromState{r})
	if err != nil {
		return err
	}

	token, err := a.githubConfig(app).Exchange(context.TODO(), r.URL.Query().Get("code"), oauth2.SetAuthURLParam("state", r.URL.Query().Get("state")))
	if err != nil {
		return err
	}

	client := github.NewClient(a.githubConfig(app).Client(context.TODO(), token))
	user, _, err := client.Users.Get(context.TODO(), "")
	if err != nil {
		return err
	}

	// TODO checks like https://roadie.io/blog/avoid-leaking-github-org-data/
	// https://docs.github.com/en/developers/apps/managing-github-apps/installing-github-apps#authorizing-users-during-installation
	// https://github.com/apps/<app name>/installations/new?state=AB12t
	w.Write([]byte(user.GetEmail()))
	return nil
}

func (a *API) home(w http.ResponseWriter, r *http.Request) error {
	// loginToApp
	return nil
}
