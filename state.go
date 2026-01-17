package katti

import (
	"errors"
)

var ErrNoMatch = errors.New("No match found")

type BindVar struct {
	Key string
	Val string
}

type MatchResult struct {
	Match    string
	Rest     string
	Pluck    bool
	NoAction bool
	BindVars []BindVar
}

type Matcher = func(prev *MatchResult) error
