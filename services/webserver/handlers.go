package main

import (
	"log"
	"net/http"
	"strconv"

	"personal-website/services/webserver/cache"
	"personal-website/services/webserver/pages"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := cache.Cache.GetPosts()
	if err != nil {
		log.Printf("failed to fetch posts from cache (%v)", err)
	}
	log.Printf("Home handler: rendering with %d posts", len(posts))
	// Varies by the home-fit-* cookies and CF ignores Vary: Cookie — never shared-cache.
	w.Header().Set("Cache-Control", "private, no-cache")
	pages.Home{
		PostsFit:    readFitCookie(r, "home-fit-posts"),
		ProjectsFit: readFitCookie(r, "home-fit-projects"),
	}.View(posts).Render(r.Context(), w)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	pages.NotFound{Path: r.URL.Path}.View().Render(r.Context(), w)
}

func readFitCookie(r *http.Request, name string) int {
	c, err := r.Cookie(name)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(c.Value)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
