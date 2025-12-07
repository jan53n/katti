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

func NewMatchResult() MatchResult {
	return MatchResult{
		BindVars: make(map[string]string),
	}
}

type Matcher = func(prev *MatchResult) error

// Literal matches if the current state input has literalString as prefix, returns error otherwise
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

// Match character if within start-end range, otherwise throw NO_MATCH
func CharRange(start rune, end rune) Matcher {
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

func SingleChar(char rune) Matcher {
	return CharRange(char, char)
}

// Wraps a matcher to print debug information about its execution.
func Leak(matcher Matcher) Matcher {
	return func(prev *MatchResult) error {
		err := matcher(prev)
		fmt.Printf("(L) %#v (%#v)\n", prev, err)
		return err
	}
}

// Match one character from a character expression
func Char(expr string) Matcher {
	class := struct {
		inverted bool

		// not implemented
		case_insensitive bool

		matchers []Matcher
	}{
		inverted:         false,
		case_insensitive: false,
		matchers:         []Matcher{},
	}

	lcChar := CharRange('a', 'z')
	ucChar := CharRange('A', 'Z')
	nChar := CharRange('0', '9')

	setGroups := func(result *MatchResult) error {
		if _, ok := result.BindVars["invert"]; ok {
			class.inverted = true
		}

		if single_char, ok := result.BindVars["single_char"]; ok {
			sr, _ := utf8.DecodeRuneInString(single_char)

			if sr == utf8.RuneError {
				return fmt.Errorf("failed to match rune")
			}

			class.matchers = append(class.matchers, CharRange(sr, sr))
		}

		if char_start, startOk := result.BindVars["char_range_start"]; startOk {
			if char_end, endOk := result.BindVars["char_range_end"]; endOk {
				sr, _ := utf8.DecodeRuneInString(char_start)
				er, _ := utf8.DecodeRuneInString(char_end)

				if sr == utf8.RuneError || er == utf8.RuneError {
					return fmt.Errorf("failed to match rune")
				}

				class.matchers = append(class.matchers, CharRange(sr, er))
			}
		}

		return nil
	}

	classSeperator := SingleChar('-')

	charClass := Sequence(
		SingleChar('['),
		Optional(
			Bind("invert", SingleChar('^')),
		),
		Repeat(
			Action(
				Alternation(
					Sequence(
						Bind("char_range_start", lcChar),
						classSeperator,
						Bind("char_range_end", lcChar),
					),
					Sequence(
						Bind("char_range_start", ucChar),
						classSeperator,
						Bind("char_range_end", ucChar),
					),
					Sequence(
						Bind("char_range_start", nChar),
						classSeperator,
						Bind("char_range_end", nChar),
					),
					Bind("single_char", Alternation(lcChar, ucChar, nChar, classSeperator)),
				),
				setGroups,
			),
			false,
		),
		SingleChar(']'),
	)

	_, perr := Parse(charClass, expr)
	matcher := Alternation(class.matchers...)

	if class.inverted {
		matcher = NegativeAssert(matcher)
	}

	return func(prev *MatchResult) (err error) {
		if perr != nil {
			return perr
		} else {
			err = matcher(prev)
		}

		return err
	}
}

var EndOfInput = NegativeAssert(AnyChar)
var AnyChar = CharRange(0, '\uFFFF')

// Act on the match result, reset bind vars, return the result
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

// Repeat until NO_MATCH is thrown
func Repeat(matcher Matcher, allowEmpty bool) Matcher {
	return func(prev *MatchResult) (err error) {
		var acc strings.Builder

		for len(prev.Rest) > 0 {
			prev.Match = ""
			err = matcher(prev)

			if err == ErrNoMatch {
				if allowEmpty {
					err = nil
				}
				break
			} else if err != nil {
				return err
			}

			if _, err = acc.WriteString(prev.Match); err != nil {
				return err
			}
		}

		prev.Match = acc.String()
		return nil
	}
}

// Try to match, otherwise silence errors
func Optional(matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		err = matcher(prev)

		if err == ErrNoMatch {
			return nil
		}

		return err
	}
}

// Bind Match string to a variable
func Bind(variable string, matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		err = matcher(prev)

		if err == nil {
			prev.BindVars[variable] = prev.Match
		}
		return err
	}
}

// Match empty if matched otherwise throw error
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

// No input is consumed. Try to match the expression. If the match does not succeed,
// just return undefined and do not consume any input, otherwise consider the match failed.
func NegativeAssert(matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		t := *prev
		matchErr := matcher(prev)

		switch matchErr {
		case nil:
			return ErrNoMatch
		case ErrNoMatch:
			*prev = t
			return nil
		default:
			return matchErr
		}
	}
}

// Join the results to make up final result, pluck optionally
func Sequence(matchers ...Matcher) Matcher {
	return func(prev *MatchResult) error {
		original := *prev
		var acc strings.Builder
		returnMode := false

		for _, m := range matchers {
			prev.Match = ""

			if err := m(prev); err != nil {
				*prev = original
				return err
			}

			if !returnMode && prev.Pluck {
				acc.Reset()
				returnMode = true
			}

			if returnMode {
				if prev.Pluck {
					acc.WriteString(prev.Match)
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

// Loop until you find a match, throw error if unknown error
func Alternation(matchers ...Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		for _, matcher := range matchers {
			err = matcher(prev)
			if err == ErrNoMatch {
				continue
			} else if err == nil {
				break
			} else {
				break
			}
		}

		return err
	}
}

// Mark the result as plucked
func Pluck(matcher Matcher) Matcher {
	return func(prev *MatchResult) (err error) {
		err = matcher(prev)

		if err == nil {
			prev.Pluck = true
		}

		return err
	}
}

// Parse creates new match and executes matcher(input), returns both error and the result of matcher function call
func Parse(matcher Matcher, input string) (*MatchResult, error) {
	match := NewMatchResult()
	match.Rest = input

	err := matcher(&match)
	return &match, err
}
