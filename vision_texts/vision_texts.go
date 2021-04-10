package vision_texts

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
)

func Detect(filename string) string {
	var result string
	ctx := context.Background()

	// Creates a client.
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	defer file.Close()

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}

	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		log.Fatalf("Failed to detect labels: %v", err)
	}

	if len(annotations) == 0 {
		fmt.Println("No text found.")
		os.Exit(0)
	}

	result = annotations[0].Description

	if !strings.HasPrefix(result, "本日の運動結果") &&
		!strings.HasPrefix(result, "Today's Results") {
		fmt.Println("No images found.")
		os.Exit(0)
	}

	return result
}
