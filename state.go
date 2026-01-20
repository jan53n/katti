package katti

import (
	"errors"
)

var ErrNoMatch = errors.New("No match found")

type BindMap []BindVar

func (bm *BindMap) Set(v BindVar) {
	*bm = append(*bm, v)
}

func (bm *BindMap) Get(k string) (string, bool) {
	for _, i := range *bm {
		if k == i.Key {
			return i.Val, true
		}
	}

	return "", false
}

type BindVar struct {
	Key string
	Val string
}

type MatchResult struct {
	Match    string
	Rest     string
	Pluck    bool
	NoAction bool
	BindVars BindMap
	Thunks   []func() error
}

type Matcher = func(prev *MatchResult) error
