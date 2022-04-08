package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tosh223/rfa/search"
)

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("projectId")
	location := r.URL.Query().Get("location")
	twitterID := r.URL.Query().Get("twitterId")
	size := r.URL.Query().Get("size")

	if projectID == "" || twitterID == "" {
		log.Fatal("Parameters not found.")
	}
	if location == "" {
		location = "us"
	}
	if size == "" {
		size = "1"
	}

	var rfa search.Rfa
	rfa.ProjectID = projectID
	rfa.Location = location
	rfa.TwitterID = twitterID
	rfa.Size = size
	rfa.Search()
}
