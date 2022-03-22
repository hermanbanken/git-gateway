package api

import (
	"errors"
	"net/http"
)

type User struct {
	GitHub    bool     `json:"github_enabled"`
	GitLab    bool     `json:"gitlab_enabled"`
	BitBucket bool     `json:"bitbucket_enabled"`
	Roles     []string `json:"roles"`
}

func (a *API) User(w http.ResponseWriter, r *http.Request) error {
	// TODO return the settings of the logged in user

	// ctx := r.Context()
	// config := getConfig(ctx)
	// settings := Settings{
	// 	GitHub:    config.GitHub.Repo != "",
	// 	GitLab:    config.GitLab.Repo != "",
	// 	BitBucket: config.BitBucket.Repo != "",
	// 	Roles:     config.Roles,
	// }
	// return sendJSON(w, http.StatusOK, &settings)

	return errors.New("not implemented")
}