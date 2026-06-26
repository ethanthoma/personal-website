package main

import (
	"log"
	"net/http"

	"personal-website/services/webserver/cache"
	"personal-website/services/webserver/pages"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := cache.Cache.GetPosts()
	if err != nil {
		log.Printf("failed to fetch posts from cache (%v)", err)
	}
	log.Printf("Home handler: rendering with %d posts", len(posts))
	w.Header().Set("Cache-Control", pageCacheControl)
	pages.Home{}.View(posts).Render(r.Context(), w)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	pages.NotFound{Path: r.URL.Path}.View().Render(r.Context(), w)
}
