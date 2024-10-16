package components

import (
	"html/template"
)

type Props struct {
	Ascii [][]string
}

func (props Props) Component(t *template.Template) error {
	name := "ascii"

	filepath := "cmd/webserver/components/" + name + "/" + name + ".tmpl"

	if _, err := t.New(name + "-component").ParseFiles(filepath); err != nil {
		return err
	} else {
		return nil
	}
}
