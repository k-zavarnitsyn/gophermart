package templates

import (
	"embed"
	"html/template"
)

const Path = "files/*.tmpl"

//go:embed files
var templatesFs embed.FS

type Loader struct {
	tpl *template.Template
}

func NewTemplateLoader(patterns ...string) (*Loader, error) {
	tpl, err := template.ParseFS(templatesFs, patterns...)
	if err != nil {
		return nil, err
	}

	return &Loader{
		tpl: tpl,
	}, nil
}

func (t *Loader) Get(path string) *template.Template {
	return t.tpl.Lookup(path)
}
