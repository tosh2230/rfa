package bq

import (
	"testing"

	"github.com/pkg/errors"
)

func TestReplaceLines(t *testing.T) {
	before := []string{
		"ジョギング",
		"Om(19809m)",
		"おなか押しにみひねり",
		"320(486回)",
		"バンザイスクワフット",
		"- 5回(5回)",
		"モモあげ",
		"Im(801m)",
		"ジョギング",
		"2 m(19809m)",
	}
	wants := []string{
		"ジョギング",
		"0m(19809m)",
		"おなか押しこみひねり",
		"32回(486回)",
		"バンザイスクワット",
		"5回(5回)",
		"モモあげ",
		"1m(801m)",
		"ジョギング",
		"2m(19809m)",
	}
	after := replaceLines(before)
	for idx, out := range after {
		want := wants[idx]
		if out != want {
			err := errors.Errorf("Fail.\n[out ]: %s\n[want]: %s", out, want)
			t.Error(err)
		}
	}
}
