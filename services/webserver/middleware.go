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
	// Do not commit headers to the underlying writer yet. The middlewareCache
	// needs to add ETag (and possibly short-circuit to 304) before the
	// status line is flushed.
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

// CSP keeps script/style/font/img/connect to 'self' (everything ships from
// /public/). Inline scripts use 'unsafe-inline' rather than per-block hashes
// because Cloudflare Web Analytics injects an inline beacon at the edge whose
// content rotates over time — pinning a hash would break navigation every
// time CF updates the snippet. 'unsafe-eval' is required by datastar's
// data-on:click expression model, which evaluates handler strings via
// new Function() at click time. Inline styles use per-block sha256s with no
// 'unsafe-inline' fallback: CF doesn't inject styles and ours are stable, so
// hash-only is genuinely strict (modern browsers ignore 'unsafe-inline' when
// hashes are present anyway, leaving it in just produced a console warning).
// frame-ancestors + X-Frame-Options together block clickjacking on both
// modern and legacy browsers. Permissions-Policy denies sensor/payment/
// clipboard APIs we don't use, including the deprecated FLoC interest-cohort
// cohort. HSTS is intentionally omitted because Cloudflare sets it at the
// edge.
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

// etagMatch performs weak comparison per RFC 7232 §2.3.2: the opaque-tag
// bodies must match, ignoring W/ prefix. Also handles comma-separated lists
// (If-None-Match: W/"a", W/"b") and the wildcard "*".
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
