package bq

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const datasetID string = "rfa"

type latest struct {
	MaxCreatedAt time.Time
}

func LoadCsv(projectID string, filename string) error {
	var tableID string = strings.ReplaceAll(filepath.Base(filename), filepath.Ext(filename), "")

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

func Query(projectID string, location string, w io.Writer) ([]latest, error) {
	var results []latest
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return results, fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	q := client.Query(
		"SELECT name FROM `bigquery-public-data.usa_names.usa_1910_2013` " +
			"WHERE state = \"TX\" " +
			"LIMIT 100")
	q.Location = "US"
	job, err := q.Run(ctx)
	if err != nil {
		return results, err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return results, err
	}
	if err := status.Err(); err != nil {
		return results, err
	}
	it, _ := job.Read(ctx)
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return results, err
		}
		results = append(results, &row)
	}
	return results, nil
}
