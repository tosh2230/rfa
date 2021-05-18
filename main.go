package main

import (
	"flag"

	"github.com/tosh223/rfa/search"
)

func main() {
	projectID := flag.String("p", "", "gcp_project_id")
	location := flag.String("l", "us", "bigquery_location")
	twitterId := flag.String("u", "", "twitter_id")
	sizeStr := flag.String("s", "1", "search_size")
	flag.Parse()

	var rfa search.Rfa
	rfa.ProjectID = *projectID
	rfa.Location = *location
	rfa.TwitterID = *twitterId
	rfa.Size = *sizeStr
	rfa.Search()
}
