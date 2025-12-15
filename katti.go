package katti

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var ErrNoMatch = errors.New("No match found")

type MatchResult struct {
	Match    string
	Rest     string
	Pluck    bool
	BindVars map[string]string
}

type Matcher = func(prev *MatchResult) error

// Literal matches if the current input has literalString as a prefix.
func Literal(literalString string) Matcher {
	return func(prev *MatchResult) (err error) {
		if strings.HasPrefix(prev.Rest, literalString) {
			prev.Match = literalString
			prev.Rest = prev.Rest[len(literalString):]
		} else {
			err = ErrNoMatch
		}

		return err
	}
}

// CharIn checks whether the first rune of the input lies between start and end (inclusive).
func CharIn(start rune, end rune) Matcher {
	return func(prev *MatchResult) (err error) {
		if len(prev.Rest) == 0 {
			return ErrNoMatch
		}

		r, size := utf8.DecodeRuneInString(prev.Rest)

		if r >= start && r <= end {
			prev.Match = prev.Rest[:size]
			prev.Rest = prev.Rest[size:]
		} else {
			err = ErrNoMatch
		}

		return err
	}
}

// Char is a convenience wrapper over CharIn for matching a single rune.
func Char(char rune) Matcher {
	return CharIn(char, char)
}

// Leak executes the matcher and always prints the match result and error to standard output.
func Leak(matcher Matcher) Matcher {
	return func(prev *MatchResult) error {
		err := matcher(prev)
		fmt.Printf("(L) %#v (%#v)\n", prev, err)
		return err
	}
}

var EndOfInput = NegativeAssert(AnyChar)
var AnyChar = CharIn(0, '\uFFFF')

// Action executes the matcher, then invokes a callback on the resulting MatchResult.
// Bound variables are reset after the callback.
func Action(matcher Matcher, cb func(result *MatchResult) error) Matcher {
	return func(prev *MatchResult) (err error) {
		err = matcher(prev)

		if err != nil {
			return err
		}

		err = cb(prev)
		prev.BindVars = map[string]string{}

		return err
	}
}

// RepeatRange applies the matcher repeatedly, requiring the number of matches to fall within the given bounds.
func RepeatRange(matcher Matcher, min, max int) Matcher {
	hasUpper := max != -1
	hasLower := min != -1

	return func(prev *MatchResult) (err error) {
		var acc strings.Builder
		matchCount := 0

		for len(prev.Rest) > 0 {
			if hasUpper && matchCount == max {
				break
			}

			prev.Match = ""
			err = matcher(prev)

			if err == ErrNoMatch {
				// *prev = original
				break
			} else if err != nil {
				return err
			}

			if _, err = acc.WriteString(prev.Match); err != nil {
				return err
			}

			matchCount++
		}

		if hasLower && matchCount < min {
			return ErrNoMatch
		}

		prev.Match = acc.String()
		return nil
	}
}

// Repeat repeatedly applies the matcher until ErrNoMatch is returned.
// If allowEmpty is false, at least one successful match is required.
func Repeat(matcher Matcher, allowEmpty bool) Matcher {
	if allowEmpty {
		return RepeatRange(matcher, 0, -1)
	}

	return RepeatRange(matcher, 1, -1)
}

// Optional attempts to apply the matcher; if it fails, the error is suppressed and no input is consumed.
func Optional(matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		err = matcher(prev)

		if err == ErrNoMatch {
			return nil
		}

		return err
	}
}

// Binds the matched string to a named variable.
func Bind(variable string, matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		err = matcher(prev)

		if err == nil {
			prev.BindVars[variable] = prev.Match
		}
		return err
	}
}

// PositiveAssert succeeds only if the matcher succeeds and does not consume input.
func PositiveAssert(matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		tempMatch := *prev
		err = matcher(prev)

		if err == nil {
			*prev = tempMatch
		}

		return err
	}

}

// NegativeAssert succeeds only if the matcher fails. It does not consume input and returns an error if the matcher succeeds.
func NegativeAssert(matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		t := *prev
		matchErr := matcher(prev)

		switch matchErr {
		case nil:
			*prev = t
			return ErrNoMatch
		case ErrNoMatch:
			*prev = t
			return nil
		default:
			return matchErr
		}
	}
}

// Sequence applies matchers in order and concatenates their matches.
// If any matcher produces a result with Pluck set to true, only plucked
// results are collected and all previously matched results are discarded.
func Sequence(matchers ...Matcher) Matcher {
	return func(prev *MatchResult) error {
		var acc strings.Builder
		pluckMode := false

		for _, m := range matchers {
			prev.Match = ""

			if err := m(prev); err != nil {
				return err
			}

			if !pluckMode && prev.Pluck {
				acc.Reset()
				pluckMode = true
			}

			if pluckMode {
				if prev.Pluck {
					acc.WriteString(prev.Match)
					prev.Pluck = false
				}

				continue
			}

			acc.WriteString(prev.Match)
		}

		prev.Match = acc.String()
		prev.Pluck = false
		return nil
	}
}

// Alternation tries each matcher in order and returns the first one that does not fail.
func Alternation(matchers ...Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		for _, matcher := range matchers {
			err = matcher(prev)
			if err == ErrNoMatch {
				continue
			} else {
				break
			}
		}

		return err
	}
}

// Pluck marks a MatchResult as plucked.
func Pluck(matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		err = matcher(prev)

		if err == nil {
			prev.Pluck = true
		}

		return err
	}
}

// Parse creates a new MatchResult, executes the matcher, and returns both the resulting MatchResult and any error.
func Parse(matcher Matcher, input string) (*MatchResult, error) {
	match := &MatchResult{
		BindVars: make(map[string]string),
		Rest:     input,
	}

	err := matcher(match)
	return match, err
}
