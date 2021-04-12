package bq

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const datasetID string = "rfa"

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

func Query(projectID string, w io.Writer) ([][]bigquery.Value, error) {
	var results [][]bigquery.Value
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return results, fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	q := client.Query(
		"select" +
			"	max(created_at)" +
			"from (" +
			"	select max(created_at) as created_at from rfa.summary" +
			"	union all" +
			"	select max(created_at) as created_at from rfa.details" +
			")",
	)
	q.Location = "US"
	q.UseLegacySQL = false
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
		results = append(results, row)
	}
	return results, nil
}
