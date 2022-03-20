package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/netlify/git-gateway/conf"
	"github.com/netlify/git-gateway/models"
)

func (a *API) systemProvider() (goth.Provider, string, error) {
	providerName := a.config.Registration.Provider
	url := a.config.API.Endpoint + "/auth/system/" + providerName + "/callback"
	switch providerName {
	case "github":
		return github.New(
			a.config.Registration.ClientID,
			a.config.Registration.ClientSecret,
			url), providerName, nil
	default:
		return nil, "", badRequestError("system supports just GitHub for now")
	}
}

func (a *API) Login(w http.ResponseWriter, r *http.Request) error {
	provider, providerName, err := a.systemProvider()
	if err != nil {
		return err
	}
	// try to get the user without re-authenticating
	if user, err := completeUserAuth(provider, providerName)(w, r); err == nil {
		_ = user
	} else {
		url, err := getAuthURL(provider, providerName)(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err)
			return nil
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
	return nil
}

func (a *API) LoginToInstance() {

}

func (a *API) SystemCallback(w http.ResponseWriter, r *http.Request) error {
	// providerName := chi.URLParam(r, "provider")
	// if providerName != "github" {
	// 	badRequestError("system uses GitHub for now")
	// }
	provider, providerName, err := a.systemProvider()
	if err != nil {
		return err
	}

	user, err := completeUserAuth(provider, providerName)(w, r)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s|%s", providerName, user.UserID)
	instance, err := a.db.GetInstance(id)
	if err == nil {
		instance.BaseConfig.GitHub.AccessToken = user.AccessToken
		err = a.db.UpdateInstance(instance)
	}
	if err != nil && models.IsNotFoundError(err) {
		instance = &models.Instance{
			ID:   id,
			UUID: uuid.New().String(),
			BaseConfig: &conf.Configuration{
				GitHub: conf.GitHubConfig{
					AccessToken: user.AccessToken,
					Repo:        user.UserID,
				},
			},
		}
		err = a.db.CreateInstance(instance)
	}
	if err != nil {
		return err
	}

	http.Redirect(w, r, "/manage/"+instance.UUID, http.StatusTemporaryRedirect)
	return nil
}

func (a *API) Logout(w http.ResponseWriter, r *http.Request) (err error) {
	err = gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
	return
}

func completeUserAuth(provider goth.Provider, providerName string) func(w http.ResponseWriter, r *http.Request) (goth.User, error) {
	return func(w http.ResponseWriter, r *http.Request) (goth.User, error) {
		value, err := gothic.GetFromSession(providerName, r)
		if err != nil {
			return goth.User{}, err
		}
		defer gothic.Logout(w, r)
		sess, err := provider.UnmarshalSession(value)
		if err != nil {
			return goth.User{}, err
		}

		err = validateState(r, sess)
		if err != nil {
			return goth.User{}, err
		}

		user, err := provider.FetchUser(sess)
		if err == nil {
			// user can be found with existing session data
			return user, err
		}

		params := r.URL.Query()
		if params.Encode() == "" && r.Method == "POST" {
			r.ParseForm()
			params = r.Form
		}

		// get new token and retry fetch
		_, err = sess.Authorize(provider, params)
		if err != nil {
			return goth.User{}, err
		}

		err = gothic.StoreInSession(providerName, sess.Marshal(), r, w)
		if err != nil {
			return goth.User{}, err
		}

		gu, err := provider.FetchUser(sess)
		return gu, err
	}
}

// validateState ensures that the state token param from the original
// AuthURL matches the one included in the current (callback) request.
func validateState(req *http.Request, sess goth.Session) error {
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		return err
	}

	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		return err
	}

	reqState := gothic.GetState(req)

	originalState := authURL.Query().Get("state")
	if originalState != "" && (originalState != reqState) {
		return errors.New("state token mismatch")
	}
	return nil
}

func getAuthURL(provider goth.Provider, providerName string) func(res http.ResponseWriter, req *http.Request) (string, error) {
	return func(res http.ResponseWriter, req *http.Request) (string, error) {
		sess, err := provider.BeginAuth(gothic.SetState(req))
		if err != nil {
			return "", err
		}

		url, err := sess.GetAuthURL()
		if err != nil {
			return "", err
		}

		err = gothic.StoreInSession(providerName, sess.Marshal(), req, res)

		if err != nil {
			return "", err
		}

		return url, err
	}
}
