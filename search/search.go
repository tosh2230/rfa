package search

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"context"

	"github.com/tosh223/rfa/bq"
	"github.com/tosh223/rfa/pixela"
	"github.com/tosh223/rfa/twitter"
	"github.com/tosh223/rfa/vision_texts"
	"golang.org/x/sync/errgroup"
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

func (rfa *Rfa) Search(ctx context.Context) (err error) {
	select {
	case <-ctx.Done():
		return
	default:
		size, _ := strconv.Atoi(rfa.Size)
		lastExecutedAt, err := getLastExecutedAt(rfa.ProjectID, rfa.Location, rfa.TwitterID)
		if err != nil {
			return err
		}

		twCfg, err := twitter.GetConfig(rfa.ProjectID, twitterSecretID)
		if err != nil {
			return err
		}

		rslts, err := twCfg.Search(&rfa.TwitterID, size, lastExecutedAt)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		eg, egCtx := errgroup.WithContext(ctx)
		for _, rslt := range rslts {
			r := rslt
			eg.Go(func() error {
				return worker(egCtx, r, &rfa.ProjectID, &rfa.TwitterID)
			})
		}
		if err := eg.Wait(); err != nil {
			return err
		}
	}

	return nil
}

func getLastExecutedAt(projectID string, location string, twitterId string) (lastExecutedAt time.Time, err error) {
	latest, err := bq.GetLatest(projectID, location, twitterId)
	if err != nil {
		var gerr *googleapi.Error
		if ok := errors.As(err, &gerr); ok {
			switch gerr.Code {
			case 404:
				lastExecutedAt = time.Time{}
				log.Printf("bigquery return 404 %v", err)
				err = nil
			default:
				log.Printf("Fatal %v", err)
				return
			}
		} else {
			log.Printf("Fatal %v", err)
			return
		}
	} else if latest == nil {
		lastExecutedAt = time.Time{}
	} else {
		lastExecutedAt = latest[0].CreatedAt.Timestamp
	}
	return
}

func worker(ctx context.Context, r twitter.Rslt, projectID *string, twitterId *string) error {
	// Detect images and load data to BigQuery
	eg, _ := errgroup.WithContext(ctx)

	select {
	case <-ctx.Done():
		return nil
	default:
		for _, url := range r.MediaUrlHttps {
			u := url
			eg.Go(func() error {
				return detecter(*projectID, *twitterId, r.CreatedAt, u)
			})
		}
		if err := eg.Wait(); err != nil {
			return err
		}

		// Pixela
		pxCfg, err := pixela.GetConfig(*projectID, pixelaSecretID)
		if err != nil {
			log.Println("Skip Pixela")
			return nil
		}

		_, err = pxCfg.Grow(r.CreatedAt)
		if err != nil {
			return fmt.Errorf("Error: pixela.CfgList.Grow %v", err)
		}
	}
	return nil
}

func detecter(projectID string, twitterId string, createdAt time.Time, url string) error {
	fmt.Println(url)
	file, err := twitter.GetImage(url)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	text := vision_texts.Detect(file.Name())
	if text == "" {
		return fmt.Errorf("Failed Detect Vision API")
	}

	tweetInfo := bq.TweetInfo{
		TwitterId: twitterId,
		CreatedAt: createdAt,
		ImageUrl:  url,
	}
	csvFile, err := tweetInfo.CreateCsv(text)
	if err != nil {
		return err
	}

	err = bq.LoadCsv(projectID, csvFile)
	if err != nil {
		return err
	}

	return nil
}
