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

type TweetInfo struct {
	TwitterId string    `json:"twitter_id" csv:"twitter_id"`
	CreatedAt time.Time `json:"created_at" csv:"created_at"`
	ImageUrl  string    `json:"image_url" csv:"image_url"`
}

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

func (tweetInfo *TweetInfo) CreateCsv(text string) (csvFile *os.File, err error) {
	lines := replaceLines(strings.Split(text, "\n"))
	lastWords := lines[len(lines)-2]

	switch {
	// summary
	case strings.HasSuffix(lastWords, "次へ"), strings.HasSuffix(lastWords, "Next"):
		csvFile, err = tweetInfo.createCsvSummary(lines)

	// details
	case strings.HasSuffix(lastWords, "とじる"), strings.HasSuffix(lastWords, "Close"):
		csvFile, err = tweetInfo.createCsvDetails(lines)
	}
	return
}

func (tweetInfo *TweetInfo) createCsvSummary(lines []string) (csvFile *os.File, err error) {
	var summary []*Summary

	for i, line := range lines {
		// fmt.Println(line)
		if RQuantity.MatchString(line) {
			summary, err = tweetInfo.setSummary(lines, i)
			if err != nil {
				return
			}
			break
		}
	}

	prefix := strings.ReplaceAll(
		filepath.Base(tweetInfo.ImageUrl),
		filepath.Ext(tweetInfo.ImageUrl),
		"",
	)
	csvName := fmt.Sprintf("summary_%s.csv", prefix)
	csvFile, err = ioutil.TempFile("", csvName)
	if err != nil {
		return
	}
	defer csvFile.Close()

	gocsv.MarshalFile(&summary, csvFile)
	return
}

func (tweetInfo *TweetInfo) setSummary(lines []string, i int) (summary []*Summary, err error) {
	var totalCaloriesBurned float64 = 0
	var totalDistanceRun float64 = 0

	totalTimeExcercising, err := time.ParseDuration(replaceTimeUnit(lines[i]))
	if err != nil {
		return
	}

	if strings.HasSuffix(lines[i+2], "kcal") {
		totalCaloriesSlice := RQuantity.FindAllString(lines[i+2], 1)
		if len(totalCaloriesSlice) > 0 {
			totalCaloriesBurned, err = strconv.ParseFloat(totalCaloriesSlice[0], 64)
			if err != nil {
				return
			}
		}

		totalDistanceRunSlice := RQuantity.FindAllString(lines[i+4], 1)
		if len(totalDistanceRunSlice) > 0 {
			totalDistanceRun, err = strconv.ParseFloat(totalDistanceRunSlice[0], 64)
			if err != nil {
				return
			}
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
		totalCaloriesBurned, err = strconv.ParseFloat(totalCaloriesInt+totalCaloriesFract, 64)
		if err != nil {
			return
		}
	}

	summary = append(summary, &Summary{
		TwitterId:            tweetInfo.TwitterId,
		CreatedAt:            tweetInfo.CreatedAt,
		ImageUrl:             tweetInfo.ImageUrl,
		TotalTimeExcercising: totalTimeExcercising,
		TotalCaloriesBurned:  totalCaloriesBurned,
		TotalDistanceRun:     totalDistanceRun,
	})
	return
}

func (tweetInfo *TweetInfo) createCsvDetails(lines []string) (csvFile *os.File, err error) {
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
			tweetInfo.setDetails(&details, line, lines[i+4])
			tweetInfo.setDetails(&details, lines[i+1], lines[i+3])
			tweetInfo.setDetails(&details, lines[i+2], lines[i+5])
			break
		} else if isExercise && rExercise.MatchString(line) {
			tweetInfo.setDetails(&details, line, lines[i+1])
		}

		if strings.HasPrefix(line, "R画面を撮影する") ||
			strings.HasPrefix(line, "画面を撮影する") {
			isExercise = true
		}
	}

	prefix := strings.ReplaceAll(filepath.Base(tweetInfo.ImageUrl), filepath.Ext(tweetInfo.ImageUrl), "")
	csvName := fmt.Sprintf("details_%s.csv", prefix)
	csvFile, err = ioutil.TempFile("", csvName)
	if err != nil {
		return
	}
	defer csvFile.Close()

	gocsv.MarshalFile(&details, csvFile)
	return
}

func (tweetInfo *TweetInfo) setDetails(details *[]*Details, nameLine string, quantityLine string) {
	rTotalQuantity := regexp.MustCompile(`\([0-9]+`)

	quantity, _ := strconv.Atoi(RQuantity.FindAllString(quantityLine, 1)[0])
	strTotalQuantity := rTotalQuantity.FindAllString(quantityLine, 1)
	totalQuantity, _ := strconv.Atoi(strings.Trim(strTotalQuantity[0], "("))
	*details = append(*details, &Details{
		TwitterId:     tweetInfo.TwitterId,
		CreatedAt:     tweetInfo.CreatedAt,
		ImageUrl:      tweetInfo.ImageUrl,
		ExerciseName:  nameLine,
		Quantity:      quantity,
		TotalQuantity: totalQuantity,
	})
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

func replaceLines(lines []string) (rLines []string) {
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
	return
}
