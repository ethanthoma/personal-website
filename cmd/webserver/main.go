package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"personal-website/cmd/webserver/pages"
)

var logRequests = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.URL)
	http.DefaultServeMux.ServeHTTP(w, req)
})

func main() {
	log.Println("Starting server...")

	pages.InitCache()

	http.HandleFunc("GET /healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("GET /robots.txt", staticHandler(http.Dir("static/seo")))

	renderer := pages.NewRenderer("cmd/webserver/components", "cmd/webserver/pages")
	ascii := createAscii()

	pageHome := &pages.Home{Renderer: renderer, Ascii: ascii}
	http.Handle("GET /", pageHome)
	http.Handle("GET /home", pageHome)
	http.Handle("GET /home/content", pageHome)

	pageBlog := &pages.Blog{Renderer: renderer, Ascii: ascii}
	http.Handle("GET /blog", pageBlog)
	http.Handle("GET /blog/content", pageBlog)

	pagePost := &pages.Post{Renderer: renderer, Ascii: ascii}
	http.Handle("GET /post/{slug}", pagePost)
	http.Handle("GET /post/{slug}/content", pagePost)

	pageProject := &pages.Project{Renderer: renderer, Ascii: ascii}
	http.Handle("GET /projects", pageProject)
	http.Handle("GET /projects/content", pageProject)

	pageWasm := &pages.Wasm{Renderer: renderer, Ascii: ascii}
	http.Handle("GET /wasm", pageWasm)
	http.Handle("GET /wasm/content", pageWasm)

	http.Handle("GET /static/", http.StripPrefix("/static/", staticHandler(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", logRequests))
}

func createAscii() [][]string {
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
