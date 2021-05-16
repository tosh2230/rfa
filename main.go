package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
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

var IsHttp bool = false

var Param struct {
	ProjectID string `json:"project_id"`
	Location  string `json:"location"`
	TwitterID string `json:"twitter_id"`
	Size      string `json:"size"`
}

func EntryPointHTTP(w http.ResponseWriter, r *http.Request) {
	if err := json.NewDecoder(r.Body).Decode(&Param); err != nil {
		log.Fatal(err)
	}
	if Param.ProjectID == "" {
		log.Fatal("Parameters not found.")
	}

	IsHttp = true
	main()
}

func main() {
	var Flag struct {
		ProjectID *string
		Location  *string
		TwitterID *string
		Size      *string
	}

	if IsHttp {
		Flag.ProjectID = &Param.ProjectID
		Flag.Location = &Param.Location
		Flag.TwitterID = &Param.TwitterID
		Flag.Size = &Param.Size
	} else {
		Flag.ProjectID = flag.String("p", "", "gcp_project_id")
		Flag.Location = flag.String("l", "us", "bigquery_location")
		Flag.TwitterID = flag.String("u", "", "twitter_id")
		Flag.Size = flag.String("s", "1", "search_size")
		flag.Parse()
	}

	size, _ := strconv.Atoi(*Flag.Size)
	lastExecutedAt := getLastExecutedAt(*Flag.ProjectID, *Flag.Location, *Flag.TwitterID)

	twCfg, err := twitter.GetConfig(*Flag.ProjectID, twitterSecretID)
	if err != nil {
		log.Fatal(err)
	}
	rslts := twCfg.Search(Flag.TwitterID, size, lastExecutedAt)

	wgWorker := new(sync.WaitGroup)
	for _, rslt := range rslts {
		wgWorker.Add(1)
		go func(r twitter.Rslt) {
			defer wgWorker.Done()
			worker(r, Flag.ProjectID, Flag.TwitterID)
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

	csvFile := bq.CreateCsv(twitterId, createdAt, url, text)
	if csvFile == nil {
		return
	}
	defer csvFile.Close()

	err = bq.LoadCsv(projectID, csvFile)
	if err != nil {
		log.Fatal(err)
	}
}
