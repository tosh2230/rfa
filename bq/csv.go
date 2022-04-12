package bq

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

var RQuantity *regexp.Regexp = regexp.MustCompile(`^[0-9]+`)
var RNumeric *regexp.Regexp = regexp.MustCompile(`\d+`)
var RNonNumeric *regexp.Regexp = regexp.MustCompile(`\D`)
var RNumericWithSpaceOrDot *regexp.Regexp = regexp.MustCompile(`\d+(?:\s|\.)+\d+`)

type SummaryColumn int
const (
	TotalDistanceRun SummaryColumn = iota
	TotalCaloriesBurned
	TotalTimeExcercising
)

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

/**
 * setSummary
 * 数文字列を抽出し、ドットや空白を含む数文字列を記号部で分割する
 * 逆順で配列を辿り、２要素ごとにカラムに代入する
 * 合計活動時間は hourもあるかもしれないので、特別対応
 */
func (tweetInfo *TweetInfo) setSummary(lines []string, i int) (summary []*Summary, err error) {
	var totalCaloriesBurned float64 = 0
	var totalDistanceRun float64 = 0
	var totalTimeExcercising time.Duration

	lines = lines[i:]
	// filter lines
	numerics := lines[:0]
	for _, v := range lines {
		if RNumeric.MatchString(v) {
			numerics = append(numerics, v)
		}
	}
	// for garbase collection
	for i := len(numerics); i < len(lines); i++ {
		lines[i] = ""
	}

	// separete space and dot
	f := func(r rune) bool {
		return r == '.' || r == ' '
	}
	for i := 0; i < len(numerics); i++ {
	// for i, v := range numerics {
		v := numerics[i]
		if RNumericWithSpaceOrDot.MatchString(v) {
			vs := strings.FieldsFunc(v, f)
			numerics = append(numerics[:i], append(vs, numerics[i+1:]...)...)
		}
	}

	// reverse
	for i := len(numerics)/2 - 1; i >= 0; i-- {
		opp := len(numerics) - 1 - i
		numerics[i], numerics[opp] = numerics[opp], numerics[i]
	}

	log.Println(numerics)

	var step SummaryColumn = TotalDistanceRun
	var stack string = ""
	for i, v := range numerics {
		match := RNumeric.FindAllString(v, 1)
		if len(match) == 0 {
			log.Fatal("")
		}
		if len(stack) > 0 {
			stack = match[0] + "." + stack
			switch step {
			case TotalDistanceRun:
				totalDistanceRun, err = strconv.ParseFloat(stack, 64)
				if err != nil {
					return
				}
			case TotalCaloriesBurned:
				totalCaloriesBurned, err = strconv.ParseFloat(stack, 64)
				if err != nil {
					return
				}
			// case TotalTimeExcercising:
			// 	totalTimeExcercising = time.ParseDuration(stack)
			}
			step += 1
			stack = ""
			if step == TotalTimeExcercising {
				numerics = numerics[i+1:]
				break
			}
		} else {
			stack = match[0]
		}
	}

	var strTimeExcercising string = ""
	if RNonNumeric.MatchString(numerics[0]) {
		strTimeExcercising = replaceTimeUnit(numerics[0])
	} else {
		units := []string{"s", "m", "h"}
		for i, numeric := range numerics {
			strTimeExcercising = numeric + units[i] + strTimeExcercising
		}
	}
	totalTimeExcercising, err = time.ParseDuration(strTimeExcercising)
	if err != nil {
		return
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

func (tweetInfo *TweetInfo) setDetails(details *[]*Details, nameLine string, quantityLine string) (err error){
	rTotalQuantity := regexp.MustCompile(`\([0-9]+`)

	strQuantity := RQuantity.FindAllString(quantityLine, 1)
	if len(strQuantity) == 0 {
		return fmt.Errorf("err failed parse quantity [%s]", nameLine)
	}
	quantity, _ := strconv.Atoi(strQuantity[0])
	strTotalQuantity := rTotalQuantity.FindAllString(quantityLine, 1)
	if len(strTotalQuantity) == 0 {
		return fmt.Errorf("err failed parse total quantity [%s]", nameLine)
	}
	totalQuantity, _ := strconv.Atoi(strings.Trim(strTotalQuantity[0], "("))
	*details = append(*details, &Details{
		TwitterId:     tweetInfo.TwitterId,
		CreatedAt:     tweetInfo.CreatedAt,
		ImageUrl:      tweetInfo.ImageUrl,
		ExerciseName:  nameLine,
		Quantity:      quantity,
		TotalQuantity: totalQuantity,
	})

	return nil
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
