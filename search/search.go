package search

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/tosh223/rfa/bq"
	"github.com/tosh223/rfa/pixela"
	"github.com/tosh223/rfa/twitter"
	"github.com/tosh223/rfa/vision_texts"
	"google.golang.org/api/googleapi"
)

const twitterSecretID string = "rfa"
const pixelaSecretID string = "rfa-pixela"

type Rfa struct {
	ProjectID string `json:"project_id"`
	Location  string `json:"location"`
	TwitterID string `json:"twitter_id"`
	Size      string `json:"size"`
}

func (rfa *Rfa) Search() {
	size, _ := strconv.Atoi(rfa.Size)
	lastExecutedAt := getLastExecutedAt(rfa.ProjectID, rfa.Location, rfa.TwitterID)

	twCfg, err := twitter.GetConfig(rfa.ProjectID, twitterSecretID)
	if err != nil {
		log.Fatal(err)
	}

	rslts, err := twCfg.Search(&rfa.TwitterID, size, lastExecutedAt)
	if err != nil {
		log.Fatal(err)
	}

	wgWorker := new(sync.WaitGroup)
	for _, rslt := range rslts {
		wgWorker.Add(1)
		go func(r twitter.Rslt) {
			defer wgWorker.Done()
			worker(r, &rfa.ProjectID, &rfa.TwitterID)
		}(rslt)
	}
	wgWorker.Wait()
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

func worker(r twitter.Rslt, projectID *string, twitterId *string) {
	// Detect images and load data to BigQuery
	wgMedia := new(sync.WaitGroup)
	for _, url := range r.MediaUrlHttps {
		wgMedia.Add(1)
		go func(u string) {
			defer wgMedia.Done()
			detecter(*projectID, *twitterId, r.CreatedAt, u)
		}(url)
	}
	wgMedia.Wait()

	// Pixela
	pxCfg, err := pixela.GetConfig(*projectID, pixelaSecretID)
	if err != nil {
		return
	}

	_, err = pxCfg.Grow(r.CreatedAt)
	if err != nil {
		log.Fatalln("Error: pixela.CfgList.Grow", err)
	}
}

func detecter(projectID string, twitterId string, createdAt time.Time, url string) {
	fmt.Println(url)
	file, err := twitter.GetImage(url)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	text := vision_texts.Detect(file.Name())
	if text == "" {
		return
	}

	tweetInfo := bq.TweetInfo{
		TwitterId: twitterId,
		CreatedAt: createdAt,
		ImageUrl:  url,
	}
	csvFile, err := tweetInfo.CreateCsv(text)
	if err != nil {
		log.Fatal(err)
	}

	err = bq.LoadCsv(projectID, csvFile)
	if err != nil {
		log.Fatal(err)
	}
}
