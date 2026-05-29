package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"personal-website/services/webserver/layouts"
)

// Explicit per-element selectors avoid datastar's children-iteration path,
// which triggered PatchElementsNoTargetsFound on Firefox even with all IDs
// present. SSE also keeps Cloudflare's Web Analytics beacon out of the
// response body (CF only injects into text/html).
func asFragment(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), layouts.FragmentKey, true)
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		h(rw, r.WithContext(ctx))

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(rw.status)

		body := rw.body.String()
		titleEnd := strings.Index(body, "</title>") + len("</title>")
		mainStart := strings.Index(body[titleEnd:], "<main") + titleEnd

		writePatch(w, "title", body[:titleEnd])
		writePatch(w, "#main", body[mainStart:])
	}
}

// Datastar's SSE parser splits each `data:` line on the first space as
// `<field> <value>`, so every HTML line gets re-prefixed with `elements `;
// otherwise continuations would be misread as new fields named after the
// first word of each HTML/CSS line.
func writePatch(w http.ResponseWriter, selector, html string) {
	fmt.Fprint(w, "event: datastar-patch-elements\n")
	fmt.Fprintf(w, "data: selector %s\n", selector)
	for _, line := range strings.Split(html, "\n") {
		fmt.Fprintf(w, "data: elements %s\n", line)
	}
	fmt.Fprint(w, "\n")
}
