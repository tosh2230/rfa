package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"rfa/twitter"
	"rfa/vision_texts"
	"strconv"
	"strings"
)

type Summary struct {
	TwitterId            string  `json:"twitter_id" csv:"twitter_id"`
	TotalTimeExcercising string  `json:"total_time_excercising" csv:"total_time_excercising"`
	TotalCaloriesBurned  float64 `json:"total_calories_burned" csv:"total_calories_burned"`
	TotalDistanceRun     float64 `json:"total_distance_run" csv:"total_distance_run"`
	CreatedAt            string  `json:"created_at" csv:"created_at"`
}

type Details struct {
	TwitterId     string `json:"twitter_id" csv:"twitter_id"`
	ExerciseName  string `json:"exercise_name" csv:"exercise_name"`
	Quantity      int    `json:"quantity" csv:"quantity"`
	TotalQuantity int    `json:"total_quantity" csv:"total_quantity"`
	CreatedAt     string `json:"created_at" csv:"created_at"`
}

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
				fmt.Println("No images found.")
				os.Exit(0)
			}

			createCSV(texts)
		}
	}
}

func createCSV(texts []string) {
	lines := strings.Split(texts[0], "\n")
	lastWords := lines[len(lines)-2]

	switch {
	// summary
	case strings.HasPrefix(lastWords, "次へ"), strings.HasPrefix(lastWords, "Next"):
		fmt.Println("Summary:")
		createCsvSummary(lines)

	// details
	case strings.HasPrefix(lastWords, "とじる"), strings.HasPrefix(lastWords, "Close"):
		fmt.Println("Details:")
		createCsvDetails(lines)

	default:
		fmt.Println("No images found.")
		os.Exit(0)
	}
}

func createCsvSummary(lines []string) {
	for _, line := range lines {
		fmt.Println(replaceLine(line))
	}
}

func createCsvDetails(lines []string) {
	var isOdd bool = (len(lines)%2 == 0)

	for i, line := range lines {
		r := regexp.MustCompile(`[0-9].+`)
		if isOdd && r.MatchString(lines[i]) && r.MatchString(lines[i+1]) {
			break
		} else {
			fmt.Println(replaceLine(line))
		}
	}
}

func replaceLine(line string) string {
	rline := strings.TrimSpace(strings.Trim(line, "*"))
	rline = strings.Replace(rline, "Om(", "0m(", 1)
	return rline
}
