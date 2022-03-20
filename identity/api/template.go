package api

import (
	"html/template"
	"io/fs"

	"github.com/netlify/git-gateway/static"
)

const (
	layoutsDir   = "templates/layouts"
	templatesDir = "templates"
	extension    = "/*.html"
)

var (
	templates map[string]*template.Template
)

// https://charly3pins.dev/blog/learn-how-to-use-the-embed-package-in-go-by-building-a-web-page-easily/
func LoadTemplates() error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(static.Files, templatesDir)
	if err != nil {
		return err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(static.Files, templatesDir+"/"+tmpl.Name(), layoutsDir+extension)
		if err != nil {
			return err
		}
		templates[tmpl.Name()] = pt
	}
	return nil
}
