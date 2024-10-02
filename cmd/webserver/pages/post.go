package pages

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"

	"personal-website/internal"
)

type Post struct {
	Renderer *Renderer
	Ascii    [][]string
}

func (p *Post) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Ascii       [][]string
		CurrentPage string
		Post        internal.Post
	}

	d := data{
		Ascii:       p.Ascii,
		CurrentPage: "blog",
	}

	slug := r.PathValue("slug")
	post, err := getPost(slug)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	d.Post = post

	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	err = p.Renderer.page(w, r, http.StatusOK, "post", d)
	if err != nil {
		log.Printf("Error rendering post %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getPost(slug string) (internal.Post, error) {
	post, err := Cache.GetPost(slug)
	if err != nil {
		log.Printf("Error getting post %s: %v", slug, err)
		return post, err
	}

	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("rose-pine"),
			),
		),
	)

	var buf bytes.Buffer
	err = mdRenderer.Convert([]byte(post.Content), &buf)
	if err != nil {
		log.Printf("Error parsing post %s to markdown: %v", slug, err)
		return post, err
	}

	post.HTML = template.HTML(buf.String())

	return post, nil
}
