package bq

import (
	"testing"

	"github.com/pkg/errors"
)

func TestReplaceLines(t *testing.T) {
	out := "0"
	want := "1"
	if out != want {
		err := errors.Errorf("Fail.\n[out ]: %s\n[want]: %s", out, want)
		t.Error(err)
	}
}
