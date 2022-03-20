package api

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// From https://golang.org/src/net/http/httputil/reverseproxy.go?s=2298:2359#L72
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func withError(fn func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dw := doneWriter{w, false}
		err := fn(&dw, r)
		if err != nil {
			logrus.Print(err)
			if !dw.done {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}
	}
}

type doneWriter struct {
	http.ResponseWriter
	done bool
}

func (w *doneWriter) WriteHeader(status int) {
	w.done = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *doneWriter) Write(b []byte) (int, error) {
	w.done = true
	return w.ResponseWriter.Write(b)
}

func withTemplate(w http.ResponseWriter, template string, handle func(t *template.Template) interface{}) error {
	t, ok := templates[template]
	if !ok {
		return fmt.Errorf("template %s not found", template)
	}
	data := handle(t)

	if err := t.Execute(w, data); err != nil {
		return err
	}
	return nil
}

func withScheme(url string) string {
	if strings.HasPrefix(url, "https://") {
		return url
	}
	return "https://" + url
}
