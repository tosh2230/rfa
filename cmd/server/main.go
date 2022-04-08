package main

import (
	"log"
	"net/http"
	"fmt"
	"os"

	"github.com/tosh223/rfa/search"
)

type Page struct {
  Title string
  Count int
}

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
	query := r.URL.Query()

	var projectID string = os.Getenv("GCP_PROJECT_ID")
	var twitterID string
	var location string
	var size string

	if len(query["projectId"]) > 0 {
		projectID = query["projectId"][0]
	} else if projectID == "" {
		msg := "Parameter[projectId] not found."
		fmt.Fprintf(w, msg)
		log.Fatal(msg)
		return
	}

	if len(query["twitterId"]) >0 {
		twitterID = query["twitterId"][0]
	} else {
		msg := "Parameter[twitterId] not found."
		fmt.Fprintf(w, msg)
		log.Fatal(msg)
		return
	}

	if len(query["location"]) > 0 {
		location = query["location"][0]
	} else {
		location = "us"
	}

	if len(query["size"]) > 0 {
		size = query["size"][0]
	} else {
		size = "1"
	}

	var rfa search.Rfa
	rfa.ProjectID = projectID
	rfa.Location = location
	rfa.TwitterID = twitterID
	rfa.Size = size
	rfa.Search()

	fmt.Fprintf(w, "HelloWorld")
}
