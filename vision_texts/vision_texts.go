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
	// Creates a client.
	ctx := context.Background()
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

	result := annotations[0].Description
	if contains(strings.Split(result, "\n"), "本日の運動結果") ||
		contains(strings.Split(result, "\n"), "Today's Results") {

		fmt.Printf("Detect results: %s\n", filename)
		fmt.Println(result)
		return result
	}

	return ""
}

func contains(texts []string, key string) bool {
	for _, text := range texts {
		if strings.HasPrefix(text, key) {
			return true
		}
	}
	return false
}
