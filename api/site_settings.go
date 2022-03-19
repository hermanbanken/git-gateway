package api

import "net/http"

func (a *API) SiteSettings(w http.ResponseWriter, r *http.Request) error {
	sendJSON(w, 500, "")
	return nil
}
