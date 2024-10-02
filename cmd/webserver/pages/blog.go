package pages

import (
	"log"
	"net/http"

	"personal-website/internal"
)

type Blog struct {
	Renderer *Renderer
	Ascii    [][]string
}

func (p *Blog) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Ascii       [][]string
		CurrentPage string
		Posts       []internal.Post
	}

	name := "blog"

	d := data{
		Ascii:       p.Ascii,
		CurrentPage: name,
	}

	posts, err := Cache.GetPosts()
	if err != nil {
		log.Printf("Error failed to get posts: %v", err)
	}

	d.Posts = posts

	err = p.Renderer.page(w, r, http.StatusOK, name, d)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
