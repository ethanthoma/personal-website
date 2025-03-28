package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"

	"personal-website/internal"
	"personal-website/services/webserver/cache"
	"personal-website/services/webserver/pages"
)

var (
	port = os.Getenv("WEBSERVER_PORT")
)

var logRequests = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.URL)
	http.DefaultServeMux.ServeHTTP(w, req)
})

func main() {
	log.Printf("Starting server on %s...", port)

	cache.InitCache()

	http.HandleFunc("GET /healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// pages

	navList := []string{"home", "blog", "projects"}

	pageHome := pages.Home{Pages: navList}
	http.Handle("GET /", templ.Handler(pageHome.View()))
	http.Handle("GET /home", templ.Handler(pageHome.View()))

	pageBlog := pages.Blog{Pages: navList}
	http.HandleFunc("GET /blog", func(w http.ResponseWriter, r *http.Request) {
		posts, err := cache.Cache.GetPosts()
		if err != nil {
			log.Printf("failed to fetch posts from cache (%v)", err)
		}

		pageBlog.View(posts).Render(r.Context(), w)
	})

	pagePost := pages.Post{Pages: navList}
	http.HandleFunc("GET /post/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		post, err := slugToHTML(slug)
		if err != nil {
			log.Printf("failed to get post %s from cache (%v)", slug, err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		pagePost.View(post).Render(r.Context(), w)
	})

	pageProjects := pages.Projects{Pages: navList}
	http.Handle("GET /projects", templ.Handler(pageProjects.View()))

	// static

	http.Handle("GET /public/", http.StripPrefix("/public/", staticHandler(http.Dir("public"))))
	http.Handle("GET /robots.txt", staticHandler(http.Dir("public/seo")))

	log.Fatal(http.ListenAndServe(":"+port, logRequests))
}

func slugToHTML(slug string) (internal.Post, error) {
	post, err := cache.Cache.GetPost(slug)
	if err != nil {
		log.Printf("error getting post %s (%v)", slug, err)
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
		log.Printf("error parsing post %s to markdown (%v)", slug, err)
		return post, err
	}

	post.HTML = template.HTML(buf.String())

	return post, nil
}

func staticHandler(root http.FileSystem) http.HandlerFunc {
	fileServer := http.FileServer(root)

	etagCache := make(map[string]string)
	if dir, ok := root.(http.Dir); ok {
		if err := filepath.Walk(string(dir), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				content, err := os.ReadFile(path)
				if err == nil {
					relativePath := strings.TrimPrefix(path, string(dir))
					hash := sha1.Sum(content)
					etagCache[filepath.ToSlash(relativePath)[1:]] = fmt.Sprintf(`"%x"`, hash)
				}
			}
			return nil
		}); err != nil {
			log.Fatalf("error walking %s (%v)", string(dir), err)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if etag, ok := etagCache[r.URL.Path]; ok {
			w.Header().Set("ETag", etag)

			if match := r.Header.Get("If-None-Match"); match != "" {
				if strings.Contains(match, etag) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}

		ext := strings.ToLower(filepath.Ext(r.URL.Path))
		switch ext {
		default:
			w.Header().Set("Cache-Control", "public, max-age=60, must-revalidate")
		}

		fileServer.ServeHTTP(w, r)
	}
}
