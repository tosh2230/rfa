package bq

import (
	"testing"

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
