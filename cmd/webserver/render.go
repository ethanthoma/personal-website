package main

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

func (t *Renderer) Content(w http.ResponseWriter, statusCode int, name string, data interface{}) error {
	tmpl, err := t.tmpl.Clone()
	if err != nil {
		log.Println("Failed to clone Renderer template")
	}
	tmpl.ParseFiles(t.pagesDir + "/" + name + ".tmpl")

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	return tmpl.ExecuteTemplate(w, "content", data)
}

func (t *Renderer) Page(w http.ResponseWriter, statusCode int, name string, data interface{}) error {
	tmpl, err := t.tmpl.Clone()
	if err != nil {
		log.Println("Failed to clone Renderer template")
	}
	tmpl.ParseGlob(t.pagesDir + "/" + name + ".tmpl")

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	return tmpl.ExecuteTemplate(w, "page", data)
}

func (t *Renderer) Component(w http.ResponseWriter, statusCode int, name string, data interface{}) error {
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	return t.tmpl.ExecuteTemplate(w, name, data)
}
