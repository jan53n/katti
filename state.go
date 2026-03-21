package katti

import (
	"errors"
)

var ErrNoMatch = errors.New("No match found")

type MatchResult struct {
	Match    []string
	Rest     string
	Bindings BindTable
	Thunks   []func() error
}

type Matcher = func(prev *MatchResult) error
