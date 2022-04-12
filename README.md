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

```sh
# build and serve
go build ./cmd/server
./server

# access
curl "http://localhost:8080/?projectId=<project-id>&twitterId=<username>&location=<bigquery-location>&size=<search-size>"
```

### CloudRun

```sh
gcloud run deploy
```

[サービス ID | Cloud Run のドキュメント | Google Cloud](https://cloud.google.com/run/docs/securing/service-identity?hl=ja)

1. Create service account
1. Update an existing service to have a new runtime service account
  - `$ gcloud run services udpate <cloudrun-service-name> --service-account <service account address>`
1. Deploy

#### URI: `/for/participants`
FirestoreからTwitterIDを取得し、[search.Search](./search/search.go)を実行する

### GitHub Actions
#### [Cloud Run](./.github/workflows/cloud-run.yml)
サービスアカウントを作っておく
- 権限：Cloud Run管理者、Cloud Run サービスエージェント、Cloud Build サービスエージェント

Workflow Identityを設定
```sh
export PROJECT_ID = <project-id>
export POOL_NAME = <pool-name>
export POOL_DISPLAY_NAME = <pool-display-name>
export PROVIDER_NAME = <provider-name>
export PROVIDER_DISPLAY_NAME = <provider-display-name>
export SA_EMAIL = <Service Account>
export GITHUB_REPO = <owner>/<repository>

# Create pool
gcloud iam workload-identity-pools create "$POOL_NAME" \
  --project="$PROJECT_ID" \
  --location="global" \
  --display-name="$POOL_DISPLAY_NAME"

# Save pool-id
export WORKLOAD_IDENTITY_POOL_ID=$( \
    gcloud iam workload-identity-pools describe "${POOL_NAME}" \
    --project="${PROJECT_ID}" --location="global" \
    --format="value(name)" \
)

# Create provider
gcloud iam workload-identity-pools providers create-oidc "$PROVIDER_NAME" \
  --project="$PROJECT_ID" \
  --location="global" \
  --workload-identity-pool="$POOL_NAME" \
  --display-name="$PROVIDER_DISPLAY_NAME" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.aud=assertion.aud,attribute.repository=assertion.repository" \
  --issuer-uri="https://token.actions.githubusercontent.com"

# Impersonate SA
gcloud iam service-accounts add-iam-policy-binding "${SA_EMAIL}" \
  --project="${PROJECT_ID}" \
  --role="foles/iam.workloadIdentityUser" \
  --member="principalSet:/qiam.googleapis.com/${WORKLOAD_IDENTITY_POOL_ID}/attribute.repository/${GITHUB_REPO}"
```

PUSH!

Ref
- https://cloud.google.com/blog/ja/products/identity-security/enabling-keyless-authentication-from-github-actions
- https://zenn.dev/vvakame/articles/gha-and-gcp-workload-identity
- https://zenn.dev/rince/scraps/4e3cbba78d2cd1


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
