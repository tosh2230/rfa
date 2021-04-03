// Sample vision-quickstart uses the Google Cloud Vision API to label an image.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
)

func main() {
	ctx := context.Background()

	// Creates a client.
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Sets the name of the image file to annotate.
	filename := "IMG_0710.PNG"

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer file.Close()
	image, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}

	// labels, err := client.DetectLabels(ctx, image, nil, 10)
	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		log.Fatalf("Failed to detect labels: %v", err)
	}

	// fmt.Println("Labels:")
	// for _, label := range labels {
	// 	fmt.Println(label.Description)
	// }

	if len(annotations) == 0 {
		fmt.Println("No text found.")
	} else {
		fmt.Println("Text:")
		for _, annotation := range annotations {
			fmt.Println(annotation.Description)
		}
	}
}
