package components

import (
	"html/template"

	ascii "personal-website/cmd/webserver/components/ascii"
	nav "personal-website/cmd/webserver/components/nav"
)

type Props struct {
	Ascii       [][]string
	PageCurrent string
	Pages       []string
}

func (props Props) Component(t *template.Template) error {
	name := "header"

	filepath := "cmd/webserver/components/" + name + "/" + name + ".tmpl"

	if err := (ascii.Props{
		Ascii: props.Ascii,
	}.Component(t)); err != nil {
		return err
	}

	if err := (nav.Props{
		PageCurrent: props.PageCurrent,
		Pages:       props.Pages,
	}.Component(t)); err != nil {
		return err
	}

	if _, err := t.New(name + "-component").ParseFiles(filepath); err != nil {
		return err
	} else {
		return nil
	}
}
