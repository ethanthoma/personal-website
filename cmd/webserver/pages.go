package main

import (
	"bytes"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
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
		CurrentPage: "blog",
		Post: Post{
			Title:   post.Title,
			Content: template.HTML(buf.String()),
		},
	}

	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	err = handler(w, http.StatusOK, "post", data)
	if err != nil {
		log.Printf("Error rendering post %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func staticHandler(root http.FileSystem) http.HandlerFunc {
	fileServer := http.FileServer(root)

	return func(w http.ResponseWriter, r *http.Request) {
		ext := strings.ToLower(filepath.Ext(r.URL.Path))

		switch ext {
		case ".css":
			w.Header().Set("Cache-Control", "public, max-age=1800, immutable")
		case ".js":
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		case ".jpg", ".jpeg", ".png", ".gif", ".ico":
			w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")
		default:
			w.Header().Set("Cache-Control", "max-age=0")
		}

		fileServer.ServeHTTP(w, r)
	}
}
