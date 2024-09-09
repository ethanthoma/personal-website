package main

import (
	"log"
	"net/http"
)

var c = NewRenderer("cmd/webserver/components", "cmd/webserver/pages")

var logRequests = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.URL)
	http.DefaultServeMux.ServeHTTP(w, req)
})

func main() {
	log.Println("Starting server...")

	InitCache()

	http.HandleFunc("GET /healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("GET /robots.txt", staticHandler(http.Dir("static/seo")))

	http.HandleFunc("GET /", homeHandler)
	http.HandleFunc("GET /home", homeHandler)
	http.HandleFunc("GET /home/content", homeHandler)

	http.HandleFunc("GET /blog", blogHandler)
	http.HandleFunc("GET /blog/content", blogHandler)

	http.HandleFunc("GET /post/{slug}", postHandler)
	http.HandleFunc("GET /post/{slug}/content", postHandler)

	http.Handle("GET /static/", http.StripPrefix("/static/", staticHandler(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", logRequests))
}
