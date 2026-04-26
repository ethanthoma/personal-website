package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"regexp"
	"slices"
	"sort"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	highlighting "github.com/yuin/goldmark-highlighting/v2"

	"personal-website/internal"
	"personal-website/services/webserver/cache"
	"personal-website/services/webserver/layouts"
	"personal-website/services/webserver/pages"
)

// post.templ renders the title and tldr in their own elements, so the
// markdown's leading H1 (always the post title) and leading blockquote (the
// `> tldr;` summary) get stripped from the rendered HTML to avoid duplication.
var (
	leadingH1Re         = regexp.MustCompile(`(?s)^\s*<h1[^>]*>.*?</h1>\s*`)
	leadingBlockquoteRe = regexp.MustCompile(`(?s)^\s*<blockquote[^>]*>.*?</blockquote>\s*`)
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

	// Register MIME types explicitly
	mime.AddExtensionType(".js", "text/javascript")
	mime.AddExtensionType(".css", "text/css")

	cache.InitCache()

	http.HandleFunc("GET /healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// pages

	navList := []string{"home", "resources"}

	handlerHome := func(w http.ResponseWriter, r *http.Request) {
		posts, err := cache.Cache.GetPosts()
		if err != nil {
			log.Printf("failed to fetch posts from cache (%v)", err)
		}
		log.Printf("Home handler: rendering with %d posts", len(posts))

		pages.Home{Pages: navList}.View(posts).Render(r.Context(), w)
	}

	postsListHandler := func(w http.ResponseWriter, r *http.Request) {
		posts, err := cache.Cache.GetPosts()
		if err != nil {
			log.Printf("failed to fetch posts from cache (%v)", err)
		}
		sort.SliceStable(posts, func(i, j int) bool {
			return posts[i].Date.After(posts[j].Date)
		})
		pages.PostsList(posts, len(posts)).Render(r.Context(), w)
	}

	projectsListHandler := func(w http.ResponseWriter, r *http.Request) {
		projects := slices.Clone(internal.Projects)
		sort.SliceStable(projects, func(i, j int) bool {
			return projects[i].Date.After(projects[j].Date)
		})
		pages.ProjectsList(projects, len(projects)).Render(r.Context(), w)
	}

	resourcesHandler := func(w http.ResponseWriter, r *http.Request) {
		pages.InfoRes{Pages: navList}.View().Render(r.Context(), w)
	}

	navHandlers := map[string]http.HandlerFunc{
		"home":      handlerHome,
		"resources": resourcesHandler,
	}

	http.HandleFunc("GET /{$}", handlerHome)
	http.HandleFunc("GET /home", handlerHome)
	http.HandleFunc("GET /resources", resourcesHandler)

	// /blog, /projects, /sitemap used to be dedicated pages; their content now
	// lives on /home (or in /public/seo/sitemap.xml for crawlers). 301 keeps
	// any external bookmarks working.
	for _, oldPath := range []string{"/blog", "/projects", "/sitemap"} {
		http.HandleFunc("GET "+oldPath, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/home", http.StatusMovedPermanently)
		})
	}

	notFoundHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		pages.NotFound{Pages: navList, Path: r.URL.Path}.View().Render(r.Context(), w)
	}

	postHandler := func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		post, err := slugToHTML(slug)
		if err != nil {
			log.Printf("post %s not found (%v)", slug, err)
			notFoundHandler(w, r)
			return
		}

		older, newer := postNeighbors(slug)

		pages.Post{Pages: navList}.View(post, older, newer).Render(r.Context(), w)
	}
	http.HandleFunc("GET /post/{slug}", postHandler)

	// Explicit Content-Type: fragments don't start with <!DOCTYPE html>, so
	// Go's auto-sniffer labels them text/plain, which datastar's @get() can't
	// HTML-patch and which browsers may skip the prefetch cache for.
	asFragment := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "public, max-age=60")
			ctx := context.WithValue(r.Context(), layouts.FragmentKey, true)
			h(w, r.WithContext(ctx))
		}
	}

	http.HandleFunc("GET /fragment/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		h, ok := navHandlers[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		asFragment(h)(w, r)
	})
	http.HandleFunc("GET /fragment/post/{slug}", asFragment(postHandler))
	http.HandleFunc("GET /fragment/posts", asFragment(postsListHandler))
	http.HandleFunc("GET /fragment/projects-list", asFragment(projectsListHandler))

	http.HandleFunc("GET /", notFoundHandler)

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
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	err = mdRenderer.Convert([]byte(post.Content), &buf)
	if err != nil {
		log.Printf("error parsing post %s to markdown (%v)", slug, err)
		return post, err
	}

	htmlStr := leadingH1Re.ReplaceAllString(buf.String(), "")
	if post.TLDR != "" {
		htmlStr = leadingBlockquoteRe.ReplaceAllString(htmlStr, "")
	}
	post.HTML = template.HTML(htmlStr)

	return post, nil
}

func postNeighbors(slug string) (older, newer *pages.PostNeighbor) {
	posts, err := cache.Cache.GetPosts()
	if err != nil {
		log.Printf("failed to fetch posts from cache (%v)", err)
		return nil, nil
	}
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	total := len(posts)
	for i, p := range posts {
		if p.Slug != slug {
			continue
		}
		if i+1 < total {
			o := posts[i+1]
			older = &pages.PostNeighbor{
				Slug:   o.Slug,
				Title:  o.Title,
				Number: fmt.Sprintf("%03d", total-(i+1)-1),
			}
		}
		if i > 0 {
			n := posts[i-1]
			newer = &pages.PostNeighbor{
				Slug:   n.Slug,
				Title:  n.Title,
				Number: fmt.Sprintf("%03d", total-(i-1)-1),
			}
		}
		break
	}
	return older, newer
}
