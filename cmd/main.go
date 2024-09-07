package main

import (
	"bytes"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"html/template"
	"log"
	"net/http"
	"strings"
	"webserver/internal"
)

var c = internal.NewRenderer("cmd/components", "cmd/pages")

type Context struct {
	CurrentPage string
	Pages       []string
	Blogs       []Blog
}

func page(name string) http.HandlerFunc {
	blogs, err := GetBlogTable()
	if err != nil {
		log.Printf("Failed to get blogs: %s", err)
	}

	data := Context{
		CurrentPage: name,
		Blogs:       blogs,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		isContentOnly := strings.HasSuffix(r.URL.Path, "/content")

		var err error
		if isContentOnly {
			err = c.Content(w, http.StatusOK, name, data)
		} else {
			err = c.Page(w, http.StatusOK, name, data)
		}

		if err != nil {
			log.Printf("Error rendering %s: %v", name, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func handler(urlPath string) func(http.ResponseWriter, int, string, interface{}) error {
	isContentOnly := strings.HasSuffix(urlPath, "/content")

	if isContentOnly {
		return c.Content
	} else {
		return c.Page
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	handler := handler(r.URL.Path)

	blog, err := GetBlogBySlug(slug)
	if err != nil {
		log.Printf("Error getting blog %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(),
		),
	)

	var buf bytes.Buffer
	err = mdRenderer.Convert([]byte(blog.Content), &buf)
	if err != nil {
		log.Printf("Error parsing blog %s to markdown: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	type Post = struct {
		Title   string
		Content template.HTML
	}

	type Data = struct {
		CurrentPage string
		Post        Post
	}

	data := Data{
		CurrentPage: "blog",
		Post: Post{
			Title:   blog.Title,
			Content: template.HTML(buf.String()),
		},
	}

	err = handler(w, http.StatusOK, "post", data)
	if err != nil {
		log.Printf("Error rendering blog %s: %v", slug, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

var logRequests = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.URL)
	http.DefaultServeMux.ServeHTTP(w, req)
})

func main() {
	log.Println("Starting server...")

	homeHandler := page("home")
	http.HandleFunc("GET /", homeHandler)
	http.HandleFunc("GET /home", homeHandler)
	http.HandleFunc("GET /home/content", homeHandler)

	releaseHandler := page("releases")
	http.HandleFunc("GET /releases", releaseHandler)
	http.HandleFunc("GET /releases/content", releaseHandler)

	blogHandler := page("blog")
	http.HandleFunc("GET /blog", blogHandler)
	http.HandleFunc("GET /blog/content", blogHandler)

	http.HandleFunc("GET /post/{slug}", postHandler)
	http.HandleFunc("GET /post/{slug}/content", postHandler)

	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", logRequests))
}
