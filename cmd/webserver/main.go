package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"personal-website/cmd/webserver/cache"

	// pages
	"personal-website/cmd/webserver/pages/blog"
	"personal-website/cmd/webserver/pages/home"
	"personal-website/cmd/webserver/pages/post"
	"personal-website/cmd/webserver/pages/projects"
)

var logRequests = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.URL)
	http.DefaultServeMux.ServeHTTP(w, req)
})

func main() {
	log.Println("Starting server...")

	cache.InitCache()

	http.HandleFunc("GET /healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("GET /robots.txt", staticHandler(http.Dir("static/seo")))

	// pages

	ascii := createAscii()

	pageHome := &home.Props{Pages: []string{"home", "blog", "projects"}, Ascii: ascii}
	http.Handle("GET /", pageHome)
	http.Handle("GET /home", pageHome)
	http.Handle("GET /home/content", pageHome)

	pageBlog := &blog.Props{Pages: []string{"home", "blog", "projects"}, Ascii: ascii}
	http.Handle("GET /blog", pageBlog)
	http.Handle("GET /blog/content", pageBlog)

	pagePost := &post.Props{Pages: []string{"home", "blog", "projects"}, Ascii: ascii}
	http.Handle("GET /post/{slug}", pagePost)
	http.Handle("GET /post/{slug}/content", pagePost)

	pageProjects := &projects.Props{Pages: []string{"home", "blog", "projects"}, Ascii: ascii}
	http.Handle("GET /projects", pageProjects)
	http.Handle("GET /projects/content", pageProjects)

	// static

	http.Handle("GET /public/", http.StripPrefix("/public/", staticHandler(http.Dir("cmd/webserver/public"))))
	http.Handle("GET /main.css", staticHandler(http.Dir("cmd/webserver/")))
	http.Handle("GET /layouts/", http.StripPrefix("/layouts/", staticHandler(http.Dir("cmd/webserver/layouts"))))
	http.Handle("GET /pages/", http.StripPrefix("/pages/", staticHandler(http.Dir("cmd/webserver/pages"))))
	http.Handle("GET /components/", http.StripPrefix("/components/", staticHandler(http.Dir("cmd/webserver/components"))))

	log.Fatal(http.ListenAndServe(":8080", logRequests))
}

func createAscii() [][]string {
	fileBuffer, err := os.ReadFile("cmd/webserver/public/images/ascii.txt")
	if err != nil {
		log.Printf("main: reading ascii (%v)", err)
	}

	lines := strings.Split(string(fileBuffer), "\n")

	asciiChars := make([][]string, len(lines))

	for i, line := range lines {
		chars := strings.Split(line, "")
		asciiChars[i] = chars
	}

	return asciiChars
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
