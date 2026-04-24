package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
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
