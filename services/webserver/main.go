package main

import (
	"log"
	"mime"
	"net/http"
	"os"

	"personal-website/services/webserver/cache"
	"personal-website/services/webserver/static"
)

var port = os.Getenv("WEBSERVER_PORT")

var logRequests = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.URL)
	http.DefaultServeMux.ServeHTTP(w, req)
})

func main() {
	log.Printf("Starting server on %s...", port)

	mime.AddExtensionType(".js", "text/javascript")
	mime.AddExtensionType(".css", "text/css")

	cache.InitCache()

	http.HandleFunc("GET /healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("GET /{$}", homeHandler)
	http.HandleFunc("GET /home", homeHandler)
	http.HandleFunc("GET /post/{slug}", postHandler)
	http.HandleFunc("GET /rss.xml", rssHandler)
	http.HandleFunc("GET /sitemap.xml", sitemapHandler)

	retiredPathsNowOnHome := []string{"/blog", "/projects", "/sitemap", "/resources"}
	for _, p := range retiredPathsNowOnHome {
		http.HandleFunc("GET "+p, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/home", http.StatusMovedPermanently)
		})
	}

	http.HandleFunc("GET /fragment/home", asFragment(homeHandler))
	http.HandleFunc("GET /fragment/post/{slug}", asFragment(postHandler))

	http.HandleFunc("GET /", notFoundHandler)
	http.Handle("GET /public/", static.Handler())
	http.Handle("GET /robots.txt", http.FileServer(http.Dir("public/seo")))

	log.Fatal(http.ListenAndServe(":"+port, middlewareSecurity(middlewareCache(logRequests))))
}
