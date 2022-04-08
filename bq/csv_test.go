package bq

import (
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestReplaceTimeUnit(t *testing.T) {
	in := "9分26秒"
	want := "9m26s"
	out := replaceTimeUnit(in)
	if out != want {
		err := errors.Errorf("Fail.\n[out ]: %s\n[want]: %s", out, want)
		t.Error(err)
	}
}

func TestReplaceLines(t *testing.T) {
	in_string := []string{
		"Om(19809m)",
		"320(486回)",
		"おなか押しにみひねり",
		"バンザイスクワフット",
		"- 5回(5回)",
		"2 m(19809m)",
		"Im(801m)",
	}
	wants := []string{
		"0m(19809m)",
		"32回(486回)",
		"おなか押しこみひねり",
		"バンザイスクワット",
		"5回(5回)",
		"2m(19809m)",
		"1m(801m)",
	}
	out_string := replaceLines(in_string)
	for idx, out := range out_string {
		want := wants[idx]
		if out != want {
			err := errors.Errorf("Fail.\n[out ]: %s\n[want]: %s", out, want)
			t.Error(err)
		}
	}
}

func TestCreateCsvDetails(t *testing.T) {
	tweetInfo := TweetInfo{
		TwitterId: "test",
		CreatedAt: time.Time{},
		ImageUrl:  "https://example.com",
	}
	in_lines := []string{
		"本日の運動結果", "test", "R", "画面を撮影する",
		"リングコン押しこみ",
		"611回(3558回)",
		"リングコン下押しこみ",
		"9回(47回)",
		"アームツイスト",
		"282回(611回)",
		"リングコン引っぱり",
		"6回(66回)",
		"バンザイコシフリ",
		"235回(376回)",
		"おなか押しこみスクワット",
		"1回(30)",
		"モモアゲアゲ",
		"108回(1019回)",
		"ジョギング",
		"215m(4479m)",
		"ねじり体側のポーズ",
		"96回(96回)",
		"ダッシュ",
		"160m(9175m)",
		"ワイドスクワット",
		"88回(198回)",
		"モモあげ",
		"131m(534m)",
		"バンザイモーニング",
		"66回(251回)",
		"スクワットキープ",
		"10秒(158秒)",
		"英雄2のポーズ",
		"リングコン下押しこみキープ",
		"リングコン引っぱりキープ",
		"60回(156回)",
		"4秒(52秒)",
		"モモデプッシュ",
		"22回(266回)",
		"4秒(376秒)",
		"おなか押しこみ",
		"20回(182回)",
		"カッコ内はプレイ開始からの累計値です", "とじる", " ",
	}

	_, err := tweetInfo.createCsvDetails(in_lines)
	if err != nil {
		t.Error(err)
	}
}

func TestSetSummaryB(t *testing.T) {
	tweetInfo := TweetInfo{
		TwitterId: "test",
		CreatedAt: time.Time{},
		ImageUrl:  "https://example.com",
	}
	in_lines := []string{
		"本日の運動結果",
		"R 画面を撮影する",
		"test",
		"12分13秒",
		"合計活動時間",
		"10.",
		"合計消費カロリー",
		"11kcal",
		"0.14km",
		"合計走行距離",
		"次へ",
		" ",
	}

	want := Summary{
		TwitterId: "test",
		CreatedAt: time.Time{},
		ImageUrl: "https://example.com",
		TotalTimeExcercising: time.Duration(12*time.Minute + 13*time.Second),
		TotalCaloriesBurned: 10.11,
		TotalDistanceRun: 0,
	}
	s, err := tweetInfo.setSummary(in_lines, 3)
	if err != nil {
		t.Error(err)
	}
	if s[0].TotalTimeExcercising != want.TotalTimeExcercising {
		t.Errorf("act:%v, except: %v", s[0].TotalTimeExcercising, want.TotalTimeExcercising)
	}
	if s[0].TotalCaloriesBurned != want.TotalCaloriesBurned {
		t.Errorf("act:%f, except: %f", s[0].TotalCaloriesBurned, want.TotalCaloriesBurned)
	}
	if s[0].TotalDistanceRun != want.TotalDistanceRun {
		t.Errorf("act:%f, except: %f", s[0].TotalDistanceRun, want.TotalDistanceRun)
	}
}

func TestSetSummaryC(t *testing.T) {
	tweetInfo := TweetInfo{
		TwitterId: "test",
		CreatedAt: time.Time{},
		ImageUrl:  "https://example.com",
	}
	in_lines := []string{
		"本日の運動結果",
		"R 画面を撮影する",
		"test",
		"12分13秒",
		"合計活動時間",
		"10.e",
		"合計消費カロリー",
		".11kcal",
		"0.14km",
		"合計走行距離",
		"次へ",
		" ",
	}

	want := Summary{
		TwitterId: "test",
		CreatedAt: time.Time{},
		ImageUrl: "https://example.com",
		TotalTimeExcercising: time.Duration(12*time.Minute + 13*time.Second),
		TotalCaloriesBurned: 10.11,
		TotalDistanceRun: 0,
	}
	s, err := tweetInfo.setSummary(in_lines, 3)
	if err != nil {
		t.Error(err)
	}
	if s[0].TotalTimeExcercising != want.TotalTimeExcercising {
		t.Errorf("act:%v, except: %v", s[0].TotalTimeExcercising, want.TotalTimeExcercising)
	}
	if s[0].TotalCaloriesBurned != want.TotalCaloriesBurned {
		t.Errorf("act:%f, except: %f", s[0].TotalCaloriesBurned, want.TotalCaloriesBurned)
	}
	if s[0].TotalDistanceRun != want.TotalDistanceRun {
		t.Errorf("act:%f, except: %f", s[0].TotalDistanceRun, want.TotalDistanceRun)
	}
}
