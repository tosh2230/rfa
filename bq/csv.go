package bq

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

type Summary struct {
	TwitterId            string    `json:"twitter_id" csv:"twitter_id"`
	TotalTimeExcercising string    `json:"total_time_excercising" csv:"total_time_excercising"`
	TotalCaloriesBurned  float64   `json:"total_calories_burned" csv:"total_calories_burned"`
	TotalDistanceRun     float64   `json:"total_distance_run" csv:"total_distance_run"`
	CreatedAt            time.Time `json:"created_at" csv:"created_at"`
}

type Details struct {
	TwitterId     string    `json:"twitter_id" csv:"twitter_id"`
	ExerciseName  string    `json:"exercise_name" csv:"exercise_name"`
	Quantity      int       `json:"quantity" csv:"quantity"`
	TotalQuantity int       `json:"total_quantity" csv:"total_quantity"`
	CreatedAt     time.Time `json:"created_at" csv:"created_at"`
}

func CreateCsv(twitterId string, createdAtStr string, text string) string {
	var csvName string = ""

	createdAt, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", createdAtStr)
	lines := replaceLines(strings.Split(text, "\n"))
	lastWords := lines[len(lines)-2]

	switch {
	// summary
	case strings.HasPrefix(lastWords, "次へ"), strings.HasPrefix(lastWords, "Next"):
		fmt.Println("Summary:")
		csvName = createCsvSummary(twitterId, createdAt, lines)

	// details
	case strings.HasPrefix(lastWords, "とじる"), strings.HasPrefix(lastWords, "Close"):
		fmt.Println("Details:")
		csvName = createCsvDetails(twitterId, createdAt, lines)
	}

	return csvName
}

func replaceLines(lines []string) []string {
	var rLines []string
	for _, line := range lines {
		rLine := strings.TrimSpace(strings.Trim(line, "*"))
		rLine = strings.Replace(rLine, "Om(", "0m(", 1)
		rLines = append(rLines, rLine)
	}
	return rLines
}

func createCsvSummary(twitterId string, createdAt time.Time, lines []string) string {
	var csvName string = "./csv/summary.csv"
	for _, line := range lines {
		fmt.Println(line)
	}
	return csvName
}

func createCsvDetails(twitterId string, createdAt time.Time, lines []string) string {
	var csvName string = "./csv/details.csv"
	var isEven bool = (len(lines)%2 == 0)
	details := []*Details{}

	for i, line := range lines {
		rExercise := regexp.MustCompile(`^[^0-9]+`)

		if i > 2 && !isEven &&
			rExercise.MatchString(line) &&
			rExercise.MatchString(lines[i+1]) {
			details = setDetails(details, twitterId, createdAt, line, lines[i+4])
			details = setDetails(details, twitterId, createdAt, lines[i+1], lines[i+3])
			details = setDetails(details, twitterId, createdAt, lines[i+2], lines[i+5])
			break
		} else if i > 2 && rExercise.MatchString(line) {
			details = setDetails(details, twitterId, createdAt, line, lines[i+1])
		}
	}

	csvfile, _ := os.OpenFile(csvName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer csvfile.Close()

	gocsv.MarshalFile(&details, csvfile)
	return csvName
}

func setDetails(details []*Details, twitterId string, createdAt time.Time, nameLine string, quantityLine string) []*Details {
	rQuantity := regexp.MustCompile(`^[0-9]+`)
	rTotalQuantity := regexp.MustCompile(`\([0-9]+`)

	quantity, _ := strconv.Atoi(rQuantity.FindAllString(quantityLine, 1)[0])
	strTotalQuantity := rTotalQuantity.FindAllString(quantityLine, 1)
	totalQuantity, _ := strconv.Atoi(strings.Trim(strTotalQuantity[0], "("))
	details = append(details, &Details{
		TwitterId:     twitterId,
		ExerciseName:  nameLine,
		Quantity:      quantity,
		TotalQuantity: totalQuantity,
		CreatedAt:     createdAt,
	})

	return details
}
