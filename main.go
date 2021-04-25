package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/tosh223/rfa/bq"
	"github.com/tosh223/rfa/twitter"
	"github.com/tosh223/rfa/vision_texts"

	"google.golang.org/api/googleapi"
)

func main() {
	wgSearch := new(sync.WaitGroup)
	wgMedia := new(sync.WaitGroup)

	projectID := flag.String("p", "", "gcp_project_id")
	location := flag.String("l", "us", "bigquery_location")
	twitterId := flag.String("u", "", "twitter_id")
	sizeStr := flag.String("s", "1", "search_size")
	flag.Parse()

	size, _ := strconv.Atoi(*sizeStr)
	lastExecutedAt := getLastExecutedAt(*projectID, *location, *twitterId)

	twCfg, err := twitter.GetConfig(*projectID)
	if err != nil {
		log.Fatal(err)
	}
	rslts := twCfg.Search(twitterId, size, lastExecutedAt)

	for _, rslt := range rslts {
		// Wait Group 01: Twitter Search
		wgSearch.Add(1)
		urls := rslt.MediaUrlHttps
		go func(r twitter.Rslt) {
			defer wgSearch.Done()
			for _, url := range urls {
				// Wait Group 02: Image Detection & Load csv to BigQuery
				wgMedia.Add(1)
				go func(u string) {
					defer wgMedia.Done()
					detectAndLoad(*projectID, *twitterId, r.CreatedAt, u)
				}(url)
			}
			wgMedia.Wait()
		}(rslt)
	}
	wgSearch.Wait()
}

func getLastExecutedAt(projectID string, location string, twitterId string) time.Time {
	var lastExecutedAt time.Time

	latest, err := bq.GetLatest(projectID, location, twitterId)
	if err != nil {
		var gerr *googleapi.Error
		if ok := errors.As(err, &gerr); ok {
			switch gerr.Code {
			case 404:
				lastExecutedAt = time.Time{}
			default:
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	} else if latest == nil {
		lastExecutedAt = time.Time{}
	} else {
		lastExecutedAt = latest[0].CreatedAt.Timestamp
	}
	return lastExecutedAt
}

func detectAndLoad(projectID string, twitterId string, createdAtStr string, url string) {
	fmt.Println(url)
	file := twitter.GetImage(url)
	defer os.Remove(file.Name())

	text := vision_texts.Detect(file.Name())
	if text == "" {
		return
	}

	csvFile := bq.CreateCsv(twitterId, createdAtStr, url, text)
	if csvFile == nil {
		return
	}
	defer csvFile.Close()

	err := bq.LoadCsv(projectID, csvFile)
	if err != nil {
		panic(err)
	}
}
