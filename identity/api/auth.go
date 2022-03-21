package api

import (
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"github.com/netlify/git-gateway/identity/models"
	"golang.org/x/oauth2"
)

func (a *API) loginToApp(w http.ResponseWriter, r *http.Request, app *models.App) {
	state := url.Values{"app": []string{strconv.FormatInt(app.ID, 10)}}.Encode()
	redirectURL := a.githubConfig(app).AuthCodeURL(state, oauth2.SetAuthURLParam("login", app.Owner.Login))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

type Session struct {
	AppID       int64
	AccessToken string
}

func (a *API) sessionFromRequest(r *http.Request) (*Session, error) {
	appIDStr, err := gothic.GetFromSession("app", r)
	if err != nil {
		return nil, err
	}
	accessToken, err := gothic.GetFromSession("accessToken", r)
	if err != nil {
		return nil, err
	}
	if accessToken == "" || appIDStr == "" {
		return nil, nil
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return &Session{AppID: appID, AccessToken: accessToken}, nil
}

func (a *API) withAuthentication(h func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		session, err := a.sessionFromRequest(r)
		if session == nil {
			http.Redirect(w, r, "/select-app", http.StatusTemporaryRedirect)
			return err
		}
		return h(w, r)
	}
}

func init() {
	// Required auth key
	authKey := []byte(os.Getenv("SESSION_AUTH_SECRET"))
	if len(authKey) == 0 {
		panic("Configure SESSION_SECRET to avoid cookie tampering")
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
