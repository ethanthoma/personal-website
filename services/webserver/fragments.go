package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"personal-website/services/webserver/layouts"
)

var fragmentCacheControl = func() string {
	if os.Getenv("DEV") == "1" {
		return "no-store"
	}
	return "private, max-age=300"
}()

// Explicit per-element selectors avoid datastar's children-iteration path that fired
// PatchElementsNoTargetsFound on Firefox; SSE also keeps Cloudflare's beacon out of the body.
func asFragment(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), layouts.FragmentKey, true)
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		h(rw, r.WithContext(ctx))

		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", fragmentCacheControl)
		w.WriteHeader(rw.status)

		body := rw.body.String()
		titleEnd := strings.Index(body, "</title>") + len("</title>")
		mainStart := strings.Index(body[titleEnd:], "<main") + titleEnd

		writePatch(w, "#doc-title", body[:titleEnd])
		writePatch(w, "#main", body[mainStart:])
	}
}

// Hand-rolled: the datastar Go SDK's NewSSE panics against the buffering middleware and forces no-cache.
// Datastar's SSE parser splits each `data:` line on the first space, so every HTML line is re-prefixed with `elements `.
func writePatch(w http.ResponseWriter, selector, html string) {
	fmt.Fprint(w, "event: datastar-patch-elements\n")
	fmt.Fprintf(w, "data: selector %s\n", selector)
	for _, line := range strings.Split(html, "\n") {
		fmt.Fprintf(w, "data: elements %s\n", line)
	}
	fmt.Fprint(w, "\n")
}
