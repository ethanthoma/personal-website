package layouts

import (
	"html/template"

	// Components
	footer "personal-website/cmd/webserver/components/footer"
	header "personal-website/cmd/webserver/components/header"
)

type BaseLayout struct {
	Ascii       [][]string
	Pages       []string
	PageCurrent string
}

func (props BaseLayout) Layout(t *template.Template) error {
	name := "base"

	filepath := "cmd/webserver/layouts/" + name + ".tmpl"

	if _, err := t.New(name + "-layout").ParseFiles(filepath); err != nil {
		return err
	}

	// Components
	if err := (footer.Props{}.Component(t)); err != nil {
		return err
	}

	if err := (header.Props{
		Ascii:       props.Ascii,
		PageCurrent: props.PageCurrent,
		Pages:       props.Pages,
	}.Component(t)); err != nil {
		return err
	}

	return nil
}
