package main

import (
	"log"
	"net/http"
	"webserver/internal"
)

var c = internal.NewRenderer("cmd/components", "cmd/pages")

var logRequests = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s\n", req.Method, req.URL)
	http.DefaultServeMux.ServeHTTP(w, req)
})

func main() {
	log.Println("Starting server...")

	http.HandleFunc("GET /", homeHandler)
	http.HandleFunc("GET /home", homeHandler)
	http.HandleFunc("GET /home/content", homeHandler)

	http.HandleFunc("GET /releases", releasesHandler)
	http.HandleFunc("GET /releases/content", releasesHandler)

	http.HandleFunc("GET /blog", blogHandler)
	http.HandleFunc("GET /blog/content", blogHandler)

	http.HandleFunc("GET /post/{slug}", postHandler)
	http.HandleFunc("GET /post/{slug}/content", postHandler)

	http.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", logRequests))
}
