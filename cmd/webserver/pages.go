package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

func router(urlPath string) func(http.ResponseWriter, int, string, interface{}) error {
	isContentOnly := strings.HasSuffix(urlPath, "/content")

	if isContentOnly {
		return c.Content
	} else {
		return c.Page
	}
}

type Context = struct {
	Ascii       [][]string
	CurrentPage string
	Post        Post
	Posts       []Post
}

var data = Context{
	Ascii: func() [][]string {
		fileBuffer, err := os.ReadFile("static/images/ascii.txt")
		if err != nil {
			log.Fatalf("Error reading ascii: %v", err)
		}

		lines := strings.Split(string(fileBuffer), "\n")

		asciiChars := make([][]string, len(lines))

		for i, line := range lines {
			chars := strings.Split(line, "")
			asciiChars[i] = chars
		}

		return asciiChars
	}(),
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	name := "home"

	data.CurrentPage = name

	err := router(r.URL.Path)(w, http.StatusOK, name, data)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
	name := "blog"

	data.CurrentPage = name

	posts, err := Cache.GetPosts()
	if err != nil {
		log.Printf("Error failed to get posts: %v", err)
	}

	data.Posts = posts

	err = router(r.URL.Path)(w, http.StatusOK, name, data)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	data.CurrentPage = "blog"

	slug := r.PathValue("slug")

	handler := router(r.URL.Path)

	post, err := Cache.GetPost(slug)
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

	post.HTML = template.HTML(buf.String())

	data.Post = post

	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	err = handler(w, http.StatusOK, "post", data)
	if err != nil {
		log.Printf("Error rendering post %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func projectsHandler(w http.ResponseWriter, r *http.Request) {
	name := "projects"

	data.CurrentPage = name

	err := router(r.URL.Path)(w, http.StatusOK, name, data)
	if err != nil {
		log.Printf("Error rendering %s: %v", name, err)
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
