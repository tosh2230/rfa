// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"fmt"
	"rfa/vision_api"
)

func main() {
	results := vision_api.Detect("img/IMG_0710.PNG")
	for _, result := range results {
		fmt.Println(result)
	}
}
