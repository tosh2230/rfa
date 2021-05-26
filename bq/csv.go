package bq

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

var RQuantity *regexp.Regexp = regexp.MustCompile(`^[0-9]+`)

type Summary struct {
	TwitterId            string        `json:"twitter_id" csv:"twitter_id"`
	CreatedAt            time.Time     `json:"created_at" csv:"created_at"`
	ImageUrl             string        `json:"image_url" csv:"image_url"`
	TotalTimeExcercising time.Duration `json:"total_time_excercising" csv:"total_time_excercising"`
	TotalCaloriesBurned  float64       `json:"total_calories_burned" csv:"total_calories_burned"`
	TotalDistanceRun     float64       `json:"total_distance_run" csv:"total_distance_run"`
}

type Details struct {
	TwitterId     string    `json:"twitter_id" csv:"twitter_id"`
	CreatedAt     time.Time `json:"created_at" csv:"created_at"`
	ImageUrl      string    `json:"image_url" csv:"image_url"`
	ExerciseName  string    `json:"exercise_name" csv:"exercise_name"`
	Quantity      int       `json:"quantity" csv:"quantity"`
	TotalQuantity int       `json:"total_quantity" csv:"total_quantity"`
}

func CreateCsv(twitterId string, createdAt time.Time, url string, text string) *os.File {
	var csvFile *os.File
	lines := replaceLines(strings.Split(text, "\n"))
	lastWords := lines[len(lines)-2]

	switch {
	// summary
	case strings.HasSuffix(lastWords, "次へ"), strings.HasSuffix(lastWords, "Next"):
		csvFile = createCsvSummary(twitterId, createdAt, url, lines)

	// details
	case strings.HasSuffix(lastWords, "とじる"), strings.HasSuffix(lastWords, "Close"):
		csvFile = createCsvDetails(twitterId, createdAt, url, lines)
	}

	return csvFile
}

func replaceTimeUnit(strTotalTime string) string {
	replaceTimeUnit := [][]string{
		{"時", "h"},
		{"分", "m"},
		{"秒", "s"},
	}

	for _, unit := range replaceTimeUnit {
		strTotalTime = strings.ReplaceAll(strTotalTime, unit[0], unit[1])
	}
	return strTotalTime
}

func replaceLines(lines []string) []string {
	var rLines []string

	replaceStr2d := [][]string{
		{"Om(", "0m("},
		{"0(", "回("},
		{"押しにみ", "押しこみ"},
		{"スクワフット", "スクワット"},
		{"- ", ""},
		{" m", "m"},
		{"Im(", "1m("},
	}

	for _, line := range lines {
		rLine := strings.TrimSpace(strings.Trim(line, "*"))
		for _, replaceStr := range replaceStr2d {
			rLine = strings.Replace(rLine, replaceStr[0], replaceStr[1], 1)
		}
		rLineSplited := strings.Split(rLine, " ")
		rLines = append(rLines, rLineSplited...)
	}
	return rLines
}

func createCsvSummary(twitterId string, createdAt time.Time, url string, lines []string) *os.File {
	prefix := strings.ReplaceAll(filepath.Base(url), filepath.Ext(url), "")
	csvName := fmt.Sprintf("summary_%s.csv", prefix)
	csvFile, err := ioutil.TempFile("", csvName)
	if err != nil {
		panic(err)
	}

	summary := setSummary(twitterId, createdAt, url, lines)
	gocsv.MarshalFile(&summary, csvFile)
	return csvFile
}

func setSummary(twitterId string, createdAt time.Time, url string, lines []string) []*Summary {
	var summary []*Summary

	for i, line := range lines {
		if RQuantity.MatchString(line) {
			return makeSummary(twitterId, createdAt, url, lines, i)
		}
	}
	return summary
}

func makeSummary(twitterId string, createdAt time.Time, url string, lines []string, i int) []*Summary {
	var summary []*Summary
	var totalCaloriesBurned float64 = 0
	var totalDistanceRun float64 = 0

	totalTimeExcercising, _ := time.ParseDuration(replaceTimeUnit(lines[i]))

	if strings.HasSuffix(lines[i+2], "kcal") {
		totalCaloriesSlice := RQuantity.FindAllString(lines[i+2], 1)
		if len(totalCaloriesSlice) > 0 {
			totalCaloriesBurned, _ = strconv.ParseFloat(totalCaloriesSlice[0], 64)
		}

		totalDistanceRunSlice := RQuantity.FindAllString(lines[i+4], 1)
		if len(totalDistanceRunSlice) > 0 {
			totalDistanceRun, _ = strconv.ParseFloat(totalDistanceRunSlice[0], 64)
		}
	} else {
		// 4分31秒
		// 合計活動時間
		// 13.
		// 合計消費力ロリー
		// 05kcal
		// 合計走行距離
		totalCaloriesInt := RQuantity.FindAllString(lines[i+2], 1)[0]
		totalCaloriesFract := RQuantity.FindAllString(lines[i+4], 1)[0]
		totalCaloriesBurned, _ = strconv.ParseFloat(totalCaloriesInt+totalCaloriesFract, 64)
	}

	return append(summary, &Summary{
		TwitterId:            twitterId,
		CreatedAt:            createdAt,
		ImageUrl:             url,
		TotalTimeExcercising: totalTimeExcercising,
		TotalCaloriesBurned:  totalCaloriesBurned,
		TotalDistanceRun:     totalDistanceRun,
	})
}

func createCsvDetails(twitterId string, createdAt time.Time, url string, lines []string) *os.File {
	var isEven bool = (len(lines)%2 == 0)
	var isExercise bool = false
	rExercise := regexp.MustCompile(`^[^0-9]+`)
	details := []*Details{}

	for i, line := range lines {
		// fmt.Println(line)
		if strings.HasPrefix(line, "カッコ内はプレイ開始からの累計値です") {
			break
		} else if isExercise && !isEven &&
			rExercise.MatchString(line) &&
			rExercise.MatchString(lines[i+1]) {
			details = setDetails(details, twitterId, createdAt, url, line, lines[i+4])
			details = setDetails(details, twitterId, createdAt, url, lines[i+1], lines[i+3])
			details = setDetails(details, twitterId, createdAt, url, lines[i+2], lines[i+5])
			break
		} else if isExercise && rExercise.MatchString(line) {
			details = setDetails(details, twitterId, createdAt, url, line, lines[i+1])
		}

		if strings.HasPrefix(line, "R画面を撮影する") ||
			strings.HasPrefix(line, "画面を撮影する") {
			isExercise = true
		}
	}

	prefix := strings.ReplaceAll(filepath.Base(url), filepath.Ext(url), "")
	csvName := fmt.Sprintf("details_%s.csv", prefix)
	csvFile, err := ioutil.TempFile("", csvName)
	if err != nil {
		panic(err)
	}

	gocsv.MarshalFile(&details, csvFile)
	return csvFile
}

func setDetails(details []*Details, twitterId string, createdAt time.Time, url string, nameLine string, quantityLine string) []*Details {
	rTotalQuantity := regexp.MustCompile(`\([0-9]+`)

	quantity, _ := strconv.Atoi(RQuantity.FindAllString(quantityLine, 1)[0])
	strTotalQuantity := rTotalQuantity.FindAllString(quantityLine, 1)
	totalQuantity, _ := strconv.Atoi(strings.Trim(strTotalQuantity[0], "("))
	details = append(details, &Details{
		TwitterId:     twitterId,
		CreatedAt:     createdAt,
		ImageUrl:      url,
		ExerciseName:  nameLine,
		Quantity:      quantity,
		TotalQuantity: totalQuantity,
	})

	return details
}
