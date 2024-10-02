package pages

import (
	"log"
	"net/http"
)

type Home struct {
	Renderer *Renderer
	Ascii    [][]string
}

func (p *Home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Ascii       [][]string
		CurrentPage string
	}

	name := "home"

	d := data{
		p.Ascii,
		name,
	}

	err := p.Renderer.page(w, r, http.StatusOK, name, d)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
