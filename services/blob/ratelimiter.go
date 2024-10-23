package main

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const duration = time.Minute

var (
	clients   = make(map[string]int)
	clientsMu = sync.Mutex{}
)

func rateLimit(next http.HandlerFunc, maxCount int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		count := clients[ip]

		if count > maxCount {
			http.Error(w, "too many requests...", http.StatusTooManyRequests)
			return
		}

		clientsMu.Lock()
		clients[ip]++
		clientsMu.Unlock()

		next.ServeHTTP(w, r)

		go func() {
			time.Sleep(duration)
			clientsMu.Lock()
			clients[ip]--
			if clients[ip] <= 0 {
				delete(clients, ip)
			}
			clientsMu.Unlock()
		}()
	}
}
