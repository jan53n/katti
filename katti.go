package katti

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

func NoAction(m Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		old := prev.NoAction
		prev.NoAction = true

		err = m(prev)

		if err == nil {
			prev.NoAction = old
		}

		return err
	}
}

func Ref(m *Matcher) Matcher {
	return func(prev *MatchResult) error {
		if m == nil || *m == nil {
			return fmt.Errorf("Ref to uninitialized matcher")
		}

		return (*m)(prev)
	}
}

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

// Char checks whether the first rune of the input equals to one of the char in []char
func Char(char ...rune) Matcher {
	return func(prev *MatchResult) (err error) {
		if len(prev.Rest) == 0 {
			return ErrNoMatch
		}

		r, size := utf8.DecodeRuneInString(prev.Rest)

		if slices.Contains(char, r) {
			prev.Match = prev.Rest[:size]
			prev.Rest = prev.Rest[size:]
			return nil
		}

		return ErrNoMatch
	}
}

// Leak executes the matcher and always prints the match result and error to standard output.
func Leak(matcher Matcher, label string) Matcher {
	return func(prev *MatchResult) error {
		t0 := time.Now()
		err := matcher(prev)

		fmt.Printf(
			"Leak (%v)\n---\nprev: %#v \nerror: (%#v)\nmatcher: %#v\ntime: %v\n---\n",
			label,
			prev,
			err,
			matcher,
			time.Since(t0),
		)

		return err
	}
}

var AnyChar = CharIn(0, utf8.MaxRune)
var EndOfInput = NegativeAssert(AnyChar)

// Action executes the matcher, then invokes a callback on the resulting MatchResult.
func Action(matcher Matcher, cb func(result *MatchResult) error) Matcher {
	return func(prev *MatchResult) (err error) {
		noAct := prev.NoAction

		err = matcher(prev)

		if err != nil {
			return err
		}

		if !noAct {
			snapshot := *prev
			prev.Thunks = append(prev.Thunks, func() error {
				return cb(&snapshot)
			})
		}

		return err
	}
}

// RepeatRange applies the matcher repeatedly, requiring the number of matches to fall within the given bounds.
func RepeatRange(matcher Matcher, min, max int, sep *Matcher) Matcher {
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

			if (matchCount+1)%2 == 0 && sep != nil {
				matcher = Sequence(*sep, matcher)
			}

			err = matcher(prev)

			if err == ErrNoMatch {
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
		return RepeatRange(matcher, 0, -1, nil)
	}

	return RepeatRange(matcher, 1, -1, nil)
}

func SepBy(matcher Matcher, sep Matcher, allowEmpty bool) Matcher {
	if allowEmpty {
		return RepeatRange(matcher, 0, -1, &sep)
	}

	return RepeatRange(matcher, 1, -1, &sep)
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
			prev.BindVars.Set(
				BindVar{
					Key: variable,
					Val: prev.Match,
				},
			)
		}
		return err
	}
}

// PositiveAssert succeeds only if the matcher succeeds and does not consume input.
func PositiveAssert(matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		t := *prev
		err = matcher(prev)

		if err == nil {
			*prev = t
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
		t := *prev
		var acc strings.Builder
		pluckMode := false

		for _, m := range matchers {
			prev.Match = ""

			if err := m(prev); err != nil {
				if err == ErrNoMatch {
					*prev = t
				}

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
		Rest: input,
	}

	err := matcher(match)

	if err != nil {
		return match, err
	}

	for _, thunk := range match.Thunks {
		if err := thunk(); err != nil {
			return match, err
		}
	}
	return match, err
}
