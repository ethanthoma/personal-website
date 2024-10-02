package pages

import (
	"log"
	"net/http"
)

type Wasm struct {
	Renderer *Renderer
	Ascii    [][]string
}

func (p *Wasm) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Ascii       [][]string
		CurrentPage string
	}

	name := "wasm"

	d := data{
		Ascii:       p.Ascii,
		CurrentPage: name,
	}

	err := p.Renderer.page(w, r, http.StatusOK, name, d)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
