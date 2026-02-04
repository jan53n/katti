package main

import (
	"testing"
)

func tryStrict(t *testing.T, exprs ...string) {
	for _, expr := range exprs {
		if mr, err := Eval(expr); err != nil {
			t.Error(err)
		} else {
			if mr.Rest != "" {
				t.Error("Rest is not empty for", expr, "with: ", mr.Rest)
			}
		}
	}
}

func TestBaseAdd(t *testing.T) {
	tryStrict(
		t,
		// "1+2",
		// "1 + 2",
		// "1 + 2 + 3 + 4",
		// "+",
		// "2",
		// "2 4 5",
		"2-3*3",
	)
}
