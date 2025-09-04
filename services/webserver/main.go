package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
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

	navList := []string{"home", "blog", "resources", "projects"}

	handlerHome := func(w http.ResponseWriter, r *http.Request) {
		posts, err := cache.Cache.GetPosts()
		if err != nil {
			log.Printf("failed to fetch posts from cache (%v)", err)
		}

		pages.Home{Pages: navList}.View(posts).Render(r.Context(), w)
	}
	http.HandleFunc("GET /", handlerHome)
	http.HandleFunc("GET /home", handlerHome)

	http.Handle("GET /blog", middlewareCache(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts, err := cache.Cache.GetPosts()
		if err != nil {
			log.Printf("failed to fetch posts from cache (%v)", err)
		}

		pages.Blog{Pages: navList}.View(posts).Render(r.Context(), w)
	})))

	http.HandleFunc("GET /post/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		post, err := slugToHTML(slug)
		if err != nil {
			log.Printf("failed to get post %s from cache (%v)", slug, err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		pages.Post{Pages: navList}.View(post).Render(r.Context(), w)
	})

	http.Handle("GET /projects", templ.Handler(pages.Projects{Pages: navList}.View()))

	http.Handle("GET /resources", templ.Handler(pages.InfoRes{Pages: navList}.View()))

	// static

	http.Handle("GET /public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	http.Handle("GET /robots.txt", http.FileServer(http.Dir("public/seo")))

	log.Fatal(http.ListenAndServe(":"+port, middlewareCache(logRequests)))
}

func slugToHTML(slug string) (internal.Post, error) {
	post, err := cache.Cache.GetPost(slug)
	if err != nil {
		log.Printf("error getting post %s from cache, trying GitHub directly (%v)", slug, err)
		post, err = internal.GetPostFromGitHub(slug)
		if err != nil {
			log.Printf("error getting post %s from GitHub (%v)", slug, err)
			return post, err
		}
	}

	styleBuilder := chroma.NewStyleBuilder("default")
	styleBuilder.AddEntry(chroma.Background, chroma.MustParseStyleEntry("bg:#ansiblack"))
	styleBuilder.AddEntry(chroma.Keyword, chroma.MustParseStyleEntry("#FFA500"))
	styleBuilder.AddEntry(chroma.Name, chroma.MustParseStyleEntry("#ansilightgray"))
	styleBuilder.AddEntry(chroma.NameVariable, chroma.MustParseStyleEntry("#A500FF"))
	styleBuilder.AddEntry(chroma.NameBuiltin, chroma.MustParseStyleEntry("#FF00A5"))
	styleBuilder.AddEntry(chroma.NameFunction, chroma.MustParseStyleEntry("#00A5FF"))
	styleBuilder.AddEntry(chroma.Literal, chroma.MustParseStyleEntry("#ansigreen"))
	styleBuilder.AddEntry(chroma.LiteralNumber, chroma.MustParseStyleEntry("#ansigreen"))
	styleBuilder.AddEntry(chroma.LiteralString, chroma.MustParseStyleEntry("#ansigreen"))

	styleBuilder.AddEntry(chroma.LineNumbers, chroma.MustParseStyleEntry("#ansidarkgray"))
	styleBuilder.AddEntry(chroma.Punctuation, chroma.MustParseStyleEntry("#a5a5a5"))
	styleBuilder.AddEntry(chroma.Generic, chroma.MustParseStyleEntry("#ansiwhite"))
	styleBuilder.AddEntry(chroma.Operator, chroma.MustParseStyleEntry("#ansiwhite"))
	styleBuilder.AddEntry(chroma.Text, chroma.MustParseStyleEntry("#ansiwhite"))

	style, err := styleBuilder.Build()

	if err != nil {
		log.Printf("error building style")
	}

	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithCustomStyle(style),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
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
