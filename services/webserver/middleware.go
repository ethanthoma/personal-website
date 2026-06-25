package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"

	"personal-website/services/webserver/highlight"
	"personal-website/services/webserver/layouts"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	// Defer the status line so middlewareCache can add ETag or short-circuit to 304.
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("http.Hijacker is not implemented by the underlying http.ResponseWriter")
	}

	return h.Hijack()
}

// script-src needs 'unsafe-inline' (Cloudflare's edge-injected beacon rotates, can't hash)
// and 'unsafe-eval' (datastar evals data-on:* via new Function()). style-src is hash-only.
// HSTS is omitted because Cloudflare sets it at the edge.
var csp = buildCSP()

const (
	cfBeaconScriptOrigin = "https://static.cloudflareinsights.com"
	cfBeaconReportOrigin = "https://cloudflareinsights.com"
)

func buildCSP() string {
	styleHashes := cspHash(layouts.FontFacesInnerCSS) + " " + cspHash(highlight.CSS)
	return "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval' " + cfBeaconScriptOrigin + "; " +
		"style-src 'self' " + styleHashes + "; " +
		"img-src 'self' data:; " +
		"font-src 'self'; " +
		"connect-src 'self' " + cfBeaconReportOrigin + "; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'; " +
		"object-src 'none'; " +
		"upgrade-insecure-requests"
}

func cspHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
}

const permissionsPolicy = "camera=(), microphone=(), geolocation=(), payment=(), usb=(), interest-cohort=()"

func middlewareSecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("Permissions-Policy", permissionsPolicy)
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func middlewareCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// ServeContent owns static validators (strong ETag, ranges); buffering would re-hash them and break ranges.
		if strings.HasPrefix(r.URL.Path, "/public/") {
			next.ServeHTTP(w, r)
			return
		}

		crw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(crw, r)

		bodyBytes := crw.body.Bytes()
		tag := eTag(bodyBytes)

		if etagMatch(r.Header.Get("If-None-Match"), tag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", tag)
		w.WriteHeader(crw.status)
		w.Write(bodyBytes)
	})
}

func eTag(content []byte) string {
	hasher := sha256.New()
	hasher.Write(content)
	hash := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("W/\"%s\"", hash)
}

// Weak comparison per RFC 7232 2.3.2: match opaque-tag bodies, ignoring the W/ prefix.
func etagMatch(ifNoneMatch, serverTag string) bool {
	if ifNoneMatch == "" {
		return false
	}
	if strings.TrimSpace(ifNoneMatch) == "*" {
		return true
	}
	server := strings.TrimPrefix(serverTag, "W/")
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		candidate = strings.TrimSpace(candidate)
		candidate = strings.TrimPrefix(candidate, "W/")
		if candidate == server {
			return true
		}
	}
	return false
}
