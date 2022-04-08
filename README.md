# rfa

Detect text in screenshots of "Ring Fit Adventure for Nintendo Switch" posted on twitter and save the play record to Google BigQuery.

## Description

![rfa](https://github.com/tosh223/rfa/blob/main/rfa.png)

## Requirements

The dependent modules are managed by Go Modules.
Please see [go.mod](https://github.com/tosh223/rfa/blob/main/go.mod).

## Preparation

- Register Twitter developer secrets to Cloud Secret Manager as `rfa`

    ```sh
    $ cat ./secret/twitter_example.json
    {
        "consumer_key": "****************",
        "consumer_secret": "****************",
        "access_token": "****************",
        "access_token_secret": "****************"
    }
    $ cp ./secret/twitter{_example,}.json  # then replace the values with your actual ones.

    $ gcloud secrets create rfa --data-file=./secret/twitter.json
    ```

- Create `rfa` dataset in Google BigQuery

    ```sh
    $ bq mk -d rfa
    ```

## Usage
### CLI

#### Build
```sh
# build
go build ./cmd/rfa

# run
./rfa --help
Usage:
  rfa [flags]

  Flags:
    -h, --help                 help for rfa
    -l, --location string      BigQuery location (default "us")
    -p, --project-id string    GCP Project ID
    -s, --search-size string   search size (default "1")
    -u, --twitter-id string    Twitter ID
```

### Web App
For CloudRun

```sh
# build and serve
go build ./cmd/server
./server

# access
curl "http://localhost:8080/?projectId=<project-id>&twitterId=<username>&location=<bigquery-location>&size=<search-size>"
```


## Test

- Set active Application Default Credentials

    ```sh
    $ gcloud auth application-default login
    ```

- Set environment variable `TEST_PROJECT_ID`

- Regist string for testing to Cloud Secret Manager as `rfa_test`

    ```sh
    $ gcloud secrets create rfa_test --data-file=./twitter/testdata/TestSetTwitterConfig.json
    ```
