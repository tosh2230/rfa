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
	"time"

	"github.com/gocarina/gocsv"
)

type Summary struct {
	TwitterId            string  `json:"twitter_id" csv:"twitter_id"`
	TotalTimeExcercising string  `json:"total_time_excercising" csv:"total_time_excercising"`
	TotalCaloriesBurned  float64 `json:"total_calories_burned" csv:"total_calories_burned"`
	TotalDistanceRun     float64 `json:"total_distance_run" csv:"total_distance_run"`
	CreatedAt            string  `json:"created_at" csv:"created_at"`
}

type Details struct {
	TwitterId     string    `json:"twitter_id" csv:"twitter_id"`
	ExerciseName  string    `json:"exercise_name" csv:"exercise_name"`
	Quantity      int       `json:"quantity" csv:"quantity"`
	TotalQuantity int       `json:"total_quantity" csv:"total_quantity"`
	CreatedAt     time.Time `json:"created_at" csv:"created_at"`
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
			fmt.Println(url)
			file := twitter.GetImage(url)
			defer os.Remove(file.Name())

			createdAt, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", rslt.CreatedAt)
			text := vision_texts.Detect(file.Name())
			createCsv(*user, createdAt, text)
		}
	}
}

func createCsv(user string, createdAt time.Time, text string) {
	lines := replaceLines(strings.Split(text, "\n"))
	lastWords := lines[len(lines)-2]

	switch {
	// summary
	case strings.HasPrefix(lastWords, "次へ"), strings.HasPrefix(lastWords, "Next"):
		fmt.Println("Summary:")
		createCsvSummary(user, createdAt, lines)

	// details
	case strings.HasPrefix(lastWords, "とじる"), strings.HasPrefix(lastWords, "Close"):
		fmt.Println("Details:")
		createCsvDetails(user, createdAt, lines)

	default:
		fmt.Println("No images found.")
		os.Exit(0)
	}
}

func replaceLines(lines []string) []string {
	var rlines []string
	for _, line := range lines {
		rline := strings.TrimSpace(strings.Trim(line, "*"))
		rline = strings.Replace(rline, "Om(", "0m(", 1)
		rlines = append(rlines, rline)
	}
	return rlines
}

func createCsvSummary(user string, createdAt time.Time, lines []string) {
	for _, line := range lines {
		fmt.Println(line)
	}
}

func createCsvDetails(user string, createdAt time.Time, lines []string) {
	var isEven bool = (len(lines)%2 == 0)
	details := []*Details{}

	for i, line := range lines {
		// fmt.Println(line)
		rExercise := regexp.MustCompile(`^[^0-9]+`)
		rQuantity := regexp.MustCompile(`^[0-9]+`)
		rTotalQuantity := regexp.MustCompile(`\([0-9]+`)

		if i > 2 && !isEven &&
			rExercise.MatchString(line) &&
			rExercise.MatchString(lines[i+1]) {
			break
		} else if i > 2 && rExercise.MatchString(line) {
			quantity, _ := strconv.Atoi(rQuantity.FindAllString(lines[i+1], 1)[0])
			strTotalQuantity := rTotalQuantity.FindAllString(lines[i+1], 1)
			totalQuantity, _ := strconv.Atoi(strings.Trim(strTotalQuantity[0], "("))
			details = append(details, &Details{
				TwitterId:     user,
				ExerciseName:  line,
				Quantity:      quantity,
				TotalQuantity: totalQuantity,
				CreatedAt:     createdAt,
			})
		}
	}
	csvStr, _ := gocsv.MarshalString(&details)
	fmt.Println(csvStr)
}
