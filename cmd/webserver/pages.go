package main

import (
	"bytes"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func router(urlPath string) func(http.ResponseWriter, int, string, interface{}) error {
	isContentOnly := strings.HasSuffix(urlPath, "/content")

	if isContentOnly {
		return c.Content
	} else {
		return c.Page
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	name := "home"

	data := struct {
		CurrentPage string
	}{
		CurrentPage: name,
	}

	err := router(r.URL.Path)(w, http.StatusOK, name, data)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func releasesHandler(w http.ResponseWriter, r *http.Request) {
	name := "releases"

	data := struct {
		CurrentPage string
	}{
		CurrentPage: name,
	}

	err := router(r.URL.Path)(w, http.StatusOK, name, data)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
	name := "blog"

	posts, err := GetPosts()
	if err != nil {
		log.Printf("Error failed to get posts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	data := struct {
		CurrentPage string
		Posts       []Post
	}{
		CurrentPage: name,
		Posts:       posts,
	}

	err = router(r.URL.Path)(w, http.StatusOK, name, data)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	handler := router(r.URL.Path)

	post, err := GetPost(slug)
	if err != nil {
		log.Printf("Error getting post %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(),
		),
	)

	var buf bytes.Buffer
	err = mdRenderer.Convert([]byte(post.Content), &buf)
	if err != nil {
		log.Printf("Error parsing post %s to markdown: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	type Post = struct {
		Title   string
		Content template.HTML
	}

	data := struct {
		CurrentPage string
		Post        Post
	}{
		CurrentPage: "post",
		Post: Post{
			Title:   post.Title,
			Content: template.HTML(buf.String()),
		},
	}

	err = handler(w, http.StatusOK, "post", data)
	if err != nil {
		log.Printf("Error rendering post %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
