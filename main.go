package main

import (
	"flag"
	"fmt"
	"os"
	"rfa/bq"
	"rfa/twitter"
	"rfa/vision_texts"
	"strconv"
)

func main() {
	projectID := flag.String("p", "", "gcp_project_id")
	twitterId := flag.String("u", "", "twitter_id")
	sizeStr := flag.String("s", "1", "search_size")
	flag.Parse()

	size, _ := strconv.Atoi(*sizeStr)
	rslts := twitter.Search(twitterId, size)

	for _, rslt := range rslts {
		fmt.Printf("@%s\n", rslt.ScreenName)

		urls := rslt.MediaUrlHttps
		for _, url := range urls {
			fmt.Println(url)
			file := twitter.GetImage(url)
			defer os.Remove(file.Name())

			text := vision_texts.Detect(file.Name())
			csvName := bq.CreateCsv(*twitterId, rslt.CreatedAt, text)
			if csvName == "" {
				continue
			}

			err := bq.LoadCsv(*projectID, csvName)
			if err != nil {
				panic(err)
			}
		}
	}
}
