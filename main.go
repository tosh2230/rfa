// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
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
	urls := twitter.Search("", 1)
	for _, url := range urls {
		file := get_image(url)
		defer os.Remove(file.Name())
		results := vision_texts.Detect(file.Name())
		for _, result := range results {
			fmt.Println(result)
		}
	}
}

func get_image(url string) *os.File {
	fmt.Println(url)
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
