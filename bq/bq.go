package bq

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const datasetID string = "rfa"

type Latest struct {
	CreatedAt time.Time `json:"created_at" csv:"created_at"`
}

func GetLatest(projectID string, location string, twitterId string) ([]*Latest, error) {
	var results []*Latest
	queryStr := fmt.Sprintf("SELECT"+
		"	MAX(created_at) AS CreatedAt"+
		" FROM ("+
		"	SELECT MAX(created_at) AS created_at FROM rfa.summary"+
		"	WHERE twitter_id = '%s'"+
		"	UNION ALL"+
		"	SELECT MAX(created_at) AS created_at FROM rfa.details"+
		"	WHERE twitter_id = '%s'"+
		")", twitterId, twitterId)

	iter, err := Query(projectID, location, queryStr)
	if err != nil {
		return nil, err
	}

	for {
		var row Latest
		err := iter.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, &row)
	}
	return results, nil
}

func Query(projectID string, location string, queryStr string) (*bigquery.RowIterator, error) {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	q := client.Query(queryStr)
	q.Location = location
	q.UseLegacySQL = false

	return q.Read(ctx)
}

func LoadCsv(projectID string, filename string) error {
	var tableID string = strings.Split(filepath.Base(filename), "_")[0]

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	source := bigquery.NewReaderSource(f)
	source.AutoDetect = true
	source.SkipLeadingRows = 1

	loader := client.Dataset(datasetID).Table(tableID).LoaderFrom(source)

	job, err := loader.Run(ctx)
	if err != nil {
		return err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	if err := status.Err(); err != nil {
		return err
	}
	return nil
}
