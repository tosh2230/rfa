package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"rfa/bq"
	"rfa/twitter"
	"rfa/vision_texts"
	"strconv"
)

func main() {
	projectID := flag.String("p", "", "gcp_project_id")
	location := flag.String("l", "us", "bigquery_location")
	twitterId := flag.String("u", "", "twitter_id")
	sizeStr := flag.String("s", "1", "search_size")
	flag.Parse()

	size, _ := strconv.Atoi(*sizeStr)

	latest, err := bq.GetLatest(*projectID, *location, *twitterId)
	if err != nil {
		log.Fatal(err)
	}
	lastExecutedAt := latest[0].CreatedAt

	rslts := twitter.Search(twitterId, size, lastExecutedAt)

	for _, rslt := range rslts {
		urls := rslt.MediaUrlHttps
		for _, url := range urls {
			fmt.Println(url)
			file := twitter.GetImage(url)
			defer os.Remove(file.Name())

			text := vision_texts.Detect(file.Name())
			if text == "" {
				continue
			}
			csvName := bq.CreateCsv(*twitterId, rslt.CreatedAt, text)
			if csvName == "" {
				continue
			}

			err := bq.LoadCsv(*projectID, csvName)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
