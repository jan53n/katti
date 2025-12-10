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

// Literal matches if the current state input has literalString as prefix
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

// CharRange checks if input's beginning character between start and end runes (start < end)
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

// SingleChar reuses CharRange to have more readability for char ranges with same start and end
func SingleChar(char rune) Matcher {
	return CharRange(char, char)
}

// Leak executes matcher and always prints the match result and error to stdin
func Leak(matcher Matcher) Matcher {
	return func(prev *MatchResult) error {
		err := matcher(prev)
		fmt.Printf("(L) %#v (%#v)\n", prev, err)
		return err
	}
}

// Char matches character class expressions, very similar to peggy.js character classes
func Char(expr string) Matcher {
	class := struct {
		inverted bool

		matchers []Matcher
	}{
		inverted: false,
		matchers: []Matcher{},
	}

	specialChar := Alternation(
		SingleChar('['),
		SingleChar(']'),
		SingleChar('-'),
		SingleChar('^'),
	)

	escapedSpecialChar := Sequence(SingleChar('\\'), Pluck(specialChar))

	nonSpecialChar := Sequence(NegativeAssert(specialChar), AnyChar)

	rangeChar := Alternation(
		escapedSpecialChar,
		nonSpecialChar,
	)

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
						Bind("char_range_start", rangeChar),
						classSeperator,
						Bind("char_range_end", rangeChar),
					),
					Bind("single_char", rangeChar),
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

// NegativeAssert will not consume input if matched
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

// NegativeAssert expects matcher to fail, does not consume input, will return error if matched
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

// Sequence joins all the match from each matchers
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

// Alternation try to find a matcher which didn't fail
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

// Pluck marks a MatchResult as “plucked.”
// When used inside a Sequence(), plucked results are given priority:
// if any matcher in the sequence is plucked, only those plucked results
// are returned; otherwise all results are returned.
//
//	Sequence(SingleChar('+'), Pluck(Literal("1")))
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
	match := &MatchResult{
		BindVars: make(map[string]string),
		Rest:     input,
	}

	err := matcher(match)
	return match, err
}
