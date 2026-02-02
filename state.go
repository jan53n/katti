package katti

import (
	"errors"
)

var ErrNoMatch = errors.New("No match found")

type BindMap []BindVar

func (bm *BindMap) Set(v BindVar) {
	*bm = append(*bm, v)
}

func (bm *BindMap) Get(k string) [][]string {
	results := [][]string{}

	for _, i := range *bm {
		if k == i.Key {
			results = append(results, i.Val)
		}
	}

	return results
}

func (bm *BindMap) Delete(k string) {
	results := []BindVar{}

	for _, i := range *bm {
		if k != i.Key {
			results = append(results, i)
		}
	}

	*bm = results
}

type BindVar struct {
	Key string
	Val []string
}

type MatchResult struct {
	Match    []string
	Rest     string
	NoAction bool
	BindVars BindMap
	Thunks   []func() error
}

type Matcher = func(prev *MatchResult) error
