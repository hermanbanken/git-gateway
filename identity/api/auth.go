package api

import (
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	// goth_github "github.com/markbates/goth/providers/github"
	"github.com/netlify/git-gateway/identity/models"
	"golang.org/x/oauth2"
)

func (a *API) loginToApp(w http.ResponseWriter, r *http.Request, app *models.App) {
	state := url.Values{"app": []string{strconv.FormatInt(app.ID, 10)}}.Encode()
	redirectURL := a.githubConfig(app).AuthCodeURL(state, oauth2.SetAuthURLParam("login", app.Owner.Login))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

type Session struct {
	AccessToken    string
	AppID          int64
	InstallationID int64
}

func (a *API) withAuthentication(h func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		// goth_github.New()
		// accessToken, err := gothic.GetFromSession("email", r)
		// accessToken, err := gothic.GetFromSession("accessToken", r)

		// loginToApp
		return h(w, r)
	}
	return nil
}

func init() {
	// Required auth key
	authKey := []byte(os.Getenv("SESSION_AUTH_SECRET"))
	if len(authKey) == 0 {
		panic("Configure SESSION_AUTH_SECRET to avoid cookie tampering")
	}
	keys := [][]byte{authKey}

	// Optional encryption key
	encKey := []byte(os.Getenv("SESSION_ENCRYPTION_SECRET"))
	if len(authKey) > 0 {
		keys = append(keys, encKey)
	}

	cookieStore := sessions.NewCookieStore(keys...)
	cookieStore.Options.HttpOnly = true
	gothic.Store = cookieStore
}

const SessionName = "_gg_session"
