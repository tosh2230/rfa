package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/tosh223/rfa/firestore"
	"github.com/tosh223/rfa/search"
	"golang.org/x/sync/errgroup"
)

type Page struct {
  Title string
  Count int
}

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)
	http.HandleFunc("/for/participants", participantHandler)

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
	ctx := context.Background()
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
		size = "15"
	}

	var rfa search.Rfa
	rfa.ProjectID = projectID
	rfa.Location = location
	rfa.TwitterID = twitterID
	rfa.Size = size
	err := rfa.Search(ctx)

	if err != nil {
		fmt.Fprintf(w, "Failed %v", err)
	} else {
		fmt.Fprintf(w, "Success")
	}

	return
}

func participantHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var projectID string = os.Getenv("GCP_PROJECT_ID")

	var rfa search.Rfa
	rfa.ProjectID = projectID
	rfa.Location = "us"
	rfa.Size = "15"

	participants, err := firestore.GetParticipants(ctx, projectID)
	if err != nil {
		msg := fmt.Sprintf("Failed %v", err)
		log.Println(msg)
		fmt.Fprintf(w, msg)
		return
	}
	log.Printf("participants: %v", participants)
	var eg errgroup.Group
	for _, v := range participants {
		rfa.TwitterID = v.ID
		if rfa.TwitterID == "" {
			fmt.Fprintf(w, "Failed getting TwitterID")
			return
		}
		rfa := rfa
		eg.Go(func() error {
			return rfa.Search(ctx)
		})
	}

	if err := eg.Wait(); err != nil {
		msg := fmt.Sprintf("Failed %v", err)
		log.Println(msg)
		fmt.Fprintf(w, msg)
	} else {
		log.Println("success")
		fmt.Fprintf(w, "Success")
	}

	return
}
