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
	"time"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"

	"personal-website/internal"
	"personal-website/services/webserver/cache"
	"personal-website/services/webserver/pages"
)

var projects = []internal.Project{
	{
		Date:  time.Date(2025, 3, 5, 0, 0, 0, 0, time.Local),
		Title: "Gleam Mobile Game",
		Description: `An Elm-based game engine written and game. 

		The engine has bindings for Three.js and uses nativescript for mobile.`,
		Lang: []string{"Gleam", "JavaScript"},
	},
	{
		Date:  time.Date(2025, 2, 11, 0, 0, 0, 0, time.Local),
		Title: "Nix builder to wrap Erlang-target Gleam code",
		Url:   "https://github.com/ethanthoma/nix-gleam-burrito",
		Description: `A Nix flake that packages Gleam projects into a single, standalone executable. 

		This builder injects a mix config into the Gleam project and uses the Burrito Elixir library to compile the code and all of its Erlang/Elixir dependencies into a self-contained binary.`,
		Lang: []string{"Nix"},
	},
	{
		Date:  time.Date(2025, 1, 25, 0, 0, 0, 0, time.Local),
		Title: "Canvas Group Quiz creation CLI in Gleam",
		Url:   "https://github.com/STASER-Lab/cgq",
		Description: `A command-line tool for the Canvas LMS that automates the creation of group-specific quizzes.
		
		Canvas doesn't allow educators to easily generate unique quizzes for different student groups, this tool does!`,
		Lang: []string{"Gleam"},
	},
	{
		Date:  time.Date(2024, 12, 10, 0, 0, 0, 0, time.Local),
		Title: "Zig native WebGPU Voxel Render",
		Url:   "https://github.com/ethanthoma/graphics",
		Description: `An exploration into voxel 3D graphics programming using Zig and WebGPU.
		
		I use Zig's comptime capabilities for compile-time generation of rendering logic and uses an instance buffer to efficiently render voxel chunks.`,
		Lang: []string{"Zig"},
	},
	{
		Date:  time.Date(2024, 9, 11, 0, 0, 0, 0, time.Local),
		Title: "Interaction Nets in Odin",
		Url:   "https://github.com/ethanthoma/interaction-net",
		Description: `A runtime for Interaction Nets, a graph-based model of computation, written in Odin.
		
		I wanted to learn the Odin language and how interaction nets work. A two for one.`,
		Lang: []string{"Odin"},
	},
	{
		Date:  time.Date(2024, 7, 8, 0, 0, 0, 0, time.Local),
		Title: "zensor: a Zig tensor library",
		Url:   "https://github.com/ethanthoma/zensor",
		Description: `A tensor library for Zig that prioritizes correctness and compile-time safety. Comes with compile-time known tensor shapes and types, eliminating a class of runtime errors tensors normally have.

		Tensor operations are converted into an AST, compiled into an intermediate representation, and then executed. It also features a work-in-progress (handmade) JIT compiler for x86.`,
		Lang: []string{"Zig"},
	},
}

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
	handlerHome := func(w http.ResponseWriter, r *http.Request) {
		posts, err := cache.Cache.GetPosts()
		if err != nil {
			log.Printf("failed to fetch posts from cache (%v)", err)
		}

		pageHome.View(posts).Render(r.Context(), w)
	}
	http.HandleFunc("GET /", handlerHome)
	http.HandleFunc("GET /home", handlerHome)

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
	http.Handle("GET /projects", templ.Handler(pageProjects.View(projects)))

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
