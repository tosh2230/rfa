package bq

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

const datasetID string = "rfa"

type Latest struct {
	CreatedAt bigquery.NullTimestamp `json:"created_at" csv:"created_at"`
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
	if iter == nil {
		return nil, nil
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

func LoadCsv(projectID string, csvFile *os.File) error {
	var tableID string = strings.Split(filepath.Base(csvFile.Name()), "_")[0]

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	f, err := os.Open(csvFile.Name())
	if err != nil {
		return err
	}
	source := bigquery.NewReaderSource(f)
	source.SourceFormat = bigquery.CSV
	source.SkipLeadingRows = 1
	source.Schema = getSchema(tableID)

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

func getSchema(tableID string) bigquery.Schema {
	var schema bigquery.Schema
	switch tableID {
	case "summary":
		schema = bigquery.Schema{
			{Name: "twitter_id", Type: bigquery.StringFieldType},
			{Name: "created_at", Type: bigquery.TimestampFieldType},
			{Name: "image_url", Type: bigquery.StringFieldType},
			{Name: "total_time_excercising", Type: bigquery.StringFieldType},
			{Name: "total_calories_burned", Type: bigquery.FloatFieldType},
			{Name: "total_distance_run", Type: bigquery.FloatFieldType},
		}
	case "details":
		schema = bigquery.Schema{
			{Name: "twitter_id", Type: bigquery.StringFieldType},
			{Name: "created_at", Type: bigquery.TimestampFieldType},
			{Name: "image_url", Type: bigquery.StringFieldType},
			{Name: "exercise_name", Type: bigquery.StringFieldType},
			{Name: "quantity", Type: bigquery.IntegerFieldType},
			{Name: "total_quantity", Type: bigquery.IntegerFieldType},
		}
	}
	return schema
}
