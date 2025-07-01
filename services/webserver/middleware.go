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
	rw.ResponseWriter.WriteHeader(code)
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
		eTag := eTag(bodyBytes)

		ifNoneMatch := strings.TrimPrefix(strings.Trim(r.Header.Get("If-None-Match"), "\""), "W/")
		contentHash := strings.TrimPrefix(eTag, "W/")

		if ifNoneMatch == strings.Trim(contentHash, "\"") {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", eTag)
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
