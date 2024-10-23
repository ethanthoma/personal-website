package main

import (
	"log"
	"net/http"
	"os"
)

const (
	Port = ":8080"
)

var (
	Token = os.Getenv("BLOB_TOKEN")
)

func main() {
	if Token == "" {
		log.Fatalf("blob-service: no token set, set BLOB_TOKEN environment var")
	}

	http.HandleFunc("DELETE /files/", BearerAuth(handlerDelete, Token))
	http.HandleFunc("GET /files/", rateLimit(handlerDownload, 100))
	http.HandleFunc("PUT /files/", BearerAuth(handlerUpload, Token))

	http.HandleFunc("GET /files", rateLimit(handlerList, 10))

	log.Printf("blob-service: server started on port %s", Port)
	err := http.ListenAndServe(Port, nil)
	if err != nil {
		log.Fatalf("blob-service: server failed (%v)", err)
	}
}
