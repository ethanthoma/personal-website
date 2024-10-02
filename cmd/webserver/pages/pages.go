package pages

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

type Renderer struct {
	pagesDir string
	tmpl     *template.Template
}

func NewRenderer(componentsDir string, pagesDir string) *Renderer {
	funcMap := template.FuncMap{
		"title": strings.Title,
		"formatDate": func(date time.Time) string {
			return date.Format("20060102")
		},
		"newRange": func(strings ...string) []string {
			return strings
		},
	}

	tmpl := template.New("").Funcs(funcMap)
	tmpl = template.Must(tmpl.ParseGlob(componentsDir + "/*.tmpl"))

	return &Renderer{
		pagesDir: pagesDir,
		tmpl:     tmpl,
	}
}

func (t *Renderer) page(w http.ResponseWriter, r *http.Request, statusCode int, name string, data interface{}) error {
	tmpl, err := t.tmpl.Clone()
	if err != nil {
		log.Println("Failed to clone Renderer template")
		return err
	}
	tmpl.ParseFiles(t.pagesDir + "/" + name + ".tmpl")

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}

	if strings.HasSuffix(r.URL.Path, "/content") {
		if tmpl.ExecuteTemplate(w, "content", data) != nil {
			log.Println("Failed to render content")
		}

		if tmpl.ExecuteTemplate(w, "oob", data) != nil {
			log.Println("Failed to render oob")
		}
	} else {
		tmpl.ParseGlob(t.pagesDir + "/" + name + ".tmpl")

		if tmpl.ExecuteTemplate(w, "base", data) != nil {
			log.Println("Failed to render base")
		}
	}

	return err
}
