// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"rfa/twitter"
	"rfa/vision_texts"
	"strings"
)

func main() {
	user := flag.String("u", "", "twitter_user_id")
	flag.Parse()

	results := twitter.Search(user, 1)
	for _, result := range results {
		url := result.MediaUrlHttps
		file := get_image(url)
		defer os.Remove(file.Name())

		texts := vision_texts.Detect(file.Name())

		fmt.Printf("@%s\n", result.ScreenName)
		fmt.Println(result.CreatedAt)
		fmt.Println(url)
		if strings.HasPrefix(texts[0], "本日") {
			fmt.Println(texts[0])
		}
	}
}

func get_image(url string) *os.File {
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	slice := strings.Split(url, "/")
	filename := fmt.Sprintf(slice[len(slice)-1])
	file, _ := ioutil.TempFile("", filename)
	io.Copy(file, response.Body)

	return file
}
