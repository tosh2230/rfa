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
var RNumericWithSpace *regexp.Regexp = regexp.MustCompile(`\d+\s+\d+`)
var RDecimal *regexp.Regexp = regexp.MustCompile(`\d+\.\d+`)
var RJapanTimeDuration *regexp.Regexp = regexp.MustCompile(`(?:\d+時間)?\d+分\d+秒`)
var RCalorie *regexp.Regexp = regexp.MustCompile(`kcal`)
var RDistance *regexp.Regexp = regexp.MustCompile(`km`)

type SummaryColumn int
const (
	TotalTimeExcercising SummaryColumn = iota
	TotalCaloriesBurned
	TotalDistanceRun
)

type TypeDetectedText int
const (
	UndefinedText = iota
	SummaryText
	DetailsText
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

	class := classificateDetectedText(text)
	switch class {
	// summary
	case SummaryText:
		csvFile, err = tweetInfo.createCsvSummary(lines)
	// details
	case DetailsText:
		csvFile, err = tweetInfo.createCsvDetails(lines)
	default:
		err = fmt.Errorf("Received Image is not expected type")
	}

	return
}

var (
	RNext = regexp.MustCompile(`(?:次へ|Next)`)
	RClose = regexp.MustCompile(`(?:とじる|Close)`)
)

func classificateDetectedText(text string) TypeDetectedText {
	switch {
	case RNext.MatchString(text):
		return SummaryText
	case RClose.MatchString(text):
		return DetailsText
	default:
		return UndefinedText
	}
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
	var totalTimeExcercising time.Duration = 0

	lines = lines[i:]
	// filter non-numeric
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
	log.Printf("summary numerics: %q", numerics)

	// 処理済みの配列要素番号
	mark_initial := len(numerics) // 初期値はMAX
	marks := map[SummaryColumn]int{
		TotalTimeExcercising: mark_initial,
		TotalCaloriesBurned: mark_initial,
		TotalDistanceRun: mark_initial,
	}
	for i, v := range numerics {
		if str := RJapanTimeDuration.FindString(v); len(str) > 0 {
			strTimeExcercising := replaceTimeUnit(str)
			totalTimeExcercising, err = time.ParseDuration(strTimeExcercising)
			if err != nil {
				return
			}
			marks[TotalTimeExcercising] = i
		} else if strDecimal := RDecimal.FindString(v); len(strDecimal) > 0 {
			DECIMAL:
			for _, w := range numerics[i:] {
				switch {
				case RCalorie.MatchString(w):
					totalCaloriesBurned, err = strconv.ParseFloat(strDecimal, 64)
					if err != nil {
						return
					}
					marks[TotalCaloriesBurned] = i
					break DECIMAL
				case RDistance.MatchString(w):
					totalDistanceRun, err = strconv.ParseFloat(strDecimal, 64)
					if err != nil {
						return
					}
					marks[TotalDistanceRun] = i
					break DECIMAL
				default:
					continue
				}
			}
		}
	}
	msgTotals := fmt.Sprintf("duration %v, %vkcal, %vkm", totalTimeExcercising, totalCaloriesBurned, totalDistanceRun)
	log.Println(msgTotals)

	remains := len(marks) // 残りの処理必要数
	for _, v := range marks {
		if v < mark_initial {
			// markがMAX以下→処理済み
			remains--
		}
		if remains == 0 {
			// if totalTimeExcercising == 0 || totalCaloriesBurned == 0 || totalDistanceRun == 0 {
			// 	msg := fmt.Sprintf("Failed parse summary: %v [marks=%v]", msgTotals, marks)
			// 	err = fmt.Errorf(msg)
			// 	return
			// }
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
	}

	// filter marks
	n := 0
	for i, v := range numerics {
		switch  {
		case i == marks[TotalTimeExcercising] || i == marks[TotalCaloriesBurned] || i >= marks[TotalDistanceRun]:
		default:
			numerics[n] = v
			n++
		}
	}
	numerics = numerics[:n]
	log.Printf("filter marks: %q", numerics)

	for i := 0; i < len(numerics); i++ {
		v := numerics[i]
		if RNumericWithSpace.MatchString(v) {
			vs := RNonNumeric.Split(v, -1)
			// filter empty string
			n := 0
			for _, v := range vs {
				if len(v) > 0 {
					vs[n] = v
					n++
				}
			}
			vs = vs[:n]

			if i+1 < len(numerics) {
				vs = append(vs, numerics[i+1:]...)
			}
			numerics = append(numerics[:i], vs...)
		}
	}

	// reverse
	for i := len(numerics)/2 - 1; i >= 0; i-- {
		opp := len(numerics) - 1 - i
		numerics[i], numerics[opp] = numerics[opp], numerics[i]
	}

	log.Printf("process stack mode: %q\n", numerics)

	var stack string = ""
	STACK:
	for i, v := range numerics {
		match := RNumeric.FindString(v)
		if len(match) == 0 {
			log.Fatal("")
		}
		if len(stack) == 0 {
			stack = match
			continue
		}
		stack = match + "." + stack
		switch {
		case totalDistanceRun == 0:
			totalDistanceRun, err = strconv.ParseFloat(stack, 64)
			if err != nil {
				return
			}
		case totalCaloriesBurned == 0:
			totalCaloriesBurned, err = strconv.ParseFloat(stack, 64)
			if err != nil {
				return
			}
		default:
			numerics = numerics[i-1:]
			break STACK
		}
		stack = ""
	}

	if totalTimeExcercising == 0 {
		var strTimeExcercising string = ""
		units := []string{"s", "m", "h"}
		for i, v := range numerics {
			strTimeExcercising = v + units[i] + strTimeExcercising
		}
		totalTimeExcercising, err = time.ParseDuration(strTimeExcercising)
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
				log.Printf("setdetail %d", i)
				tweetInfo.setDetails(&details, line, lines[i+4])
				tweetInfo.setDetails(&details, lines[i+1], lines[i+3])
				tweetInfo.setDetails(&details, lines[i+2], lines[i+5])
			break
		} else if isExercise && rExercise.MatchString(line) {
			log.Printf("setdetail %s=%s", line, lines[i+1])
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
		{"時間", "h"},
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
