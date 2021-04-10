// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"flag"
	"fmt"
	"os"
	"rfa/twitter"
	"rfa/vision_texts"
	"strconv"
	"strings"
)

func main() {
	user := flag.String("u", "", "twitter_user_id")
	sizeStr := flag.String("s", "1", "search_size")
	flag.Parse()

	size, _ := strconv.Atoi(*sizeStr)
	rslts := twitter.Search(user, size)

	for _, rslt := range rslts {
		fmt.Printf("@%s\n", rslt.ScreenName)
		fmt.Println(rslt.CreatedAt)

		urls := rslt.MediaUrlHttps
		for _, url := range urls {
			file := twitter.GetImage(url)
			defer os.Remove(file.Name())

			texts := vision_texts.Detect(file.Name())

			fmt.Println(url)
			if !strings.HasPrefix(texts[0], "本日の運動結果") &&
				!strings.HasPrefix(texts[0], "Today's Results") {
				os.Exit(0)
			}
			fmt.Println(texts[0])
		}
	}
}
