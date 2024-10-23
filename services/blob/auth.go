package main

import (
	"log"
	"net/http"
	"strings"
)

func BearerAuth(handler http.HandlerFunc, validToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			log.Printf("auth is %s", auth)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		if token != validToken {
			log.Printf("auth is %s, not %s", auth, validToken)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	}
}
