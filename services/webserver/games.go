package main

import (
	"net/http"

	"personal-website/services/webserver/pages"
	"personal-website/services/webserver/static"
)

var gameCards = []pages.GameCard{
	{Slug: "globle", Title: "Globle", Blurb: "Guess the mystery country. Warmer is closer."},
}

func gamesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", pageCacheControl)
	pages.Games{}.View(gameCards).Render(r.Context(), w)
}

// Games are standalone client-side apps, so they full-page load (no SSE fragment nav):
// datastar patches inject the shell's inline <script> without executing it.
func gameHandler(w http.ResponseWriter, r *http.Request) {
	switch r.PathValue("slug") {
	case "globle":
		w.Header().Set("Cache-Control", pageCacheControl)
		pages.Globle{}.View(static.Asset("games/countries.json")).Render(r.Context(), w)
	default:
		notFoundHandler(w, r)
	}
}
