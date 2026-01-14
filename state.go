package katti

import (
	"errors"
	"maps"
)

var ErrNoMatch = errors.New("No match found")

type MatchResult struct {
	Match    string
	Rest     string
	Pluck    bool
	NoAction bool
	BindVars map[string][]string
}

type Matcher = func(prev *MatchResult) error

func snapshot(result *MatchResult) MatchResult {
	newResult := *result
	newResult.BindVars = maps.Clone(newResult.BindVars)
	return newResult
}
