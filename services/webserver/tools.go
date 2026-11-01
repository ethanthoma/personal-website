package main

import (
	"net/http"

	"personal-website/services/webserver/pages"
	"personal-website/services/webserver/static"
)

var toolCards = []pages.ToolCard{
	{
		Slug:     "image-editor",
		Title:    "Image editor",
		Blurb:    "Paste and arrange images, sketch, draw lines, and fill magic-wand selections.",
		Category: "Create",
	},
	{
		Slug:     "globle",
		Title:    "Globle",
		Blurb:    "Guess the mystery country. Warmer is closer.",
		Category: "Play",
	},
}

func toolsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", pageCacheControl)
	pages.Tools{}.View(toolCards).Render(r.Context(), w)
}

// Tools are standalone client-side apps, so they full-page load (no SSE fragment nav):
// datastar patches inject the shell's inline <script> without executing it.
func toolHandler(w http.ResponseWriter, r *http.Request) {
	switch r.PathValue("slug") {
	case "image-editor":
		w.Header().Set("Cache-Control", pageCacheControl)
		pages.ImageEditor{}.View().Render(r.Context(), w)
	case "globle":
		w.Header().Set("Cache-Control", pageCacheControl)
		pages.Globle{}.View(static.Asset("games/countries.json")).Render(r.Context(), w)
	default:
		notFoundHandler(w, r)
	}
}
