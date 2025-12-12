package katti_test

import (
	"reflect"
	"testing"

	. "jnsn.in/katti"
)

type testTableItem struct {
	name    string
	matcher Matcher
	input   string
	result  *MatchResult
	err     error
}

func expectResult(t *testing.T, matcher Matcher, input string, expected *MatchResult, expectedErr error) {
	r, err := Parse(matcher, input)

	if err != nil {
		if expectedErr != nil {
			if expectedErr != err {
				t.Errorf("expected err '%#v' but got '%#v'", expectedErr, err)
			}
		} else {
			t.Fatalf("unexpected error: %#v", err)
		}
	}

	if expected == nil {
		return
	}

	if r.Match != expected.Match {
		t.Errorf("expected match '%v' but got '%v'", expected.Match, r.Match)
	}

	if r.Rest != expected.Rest {
		t.Errorf("expected rest '%v' but got '%v'", expected.Rest, r.Rest)
	}

	if expected.BindVars == nil {
		expected.BindVars = make(map[string]string)
	}

	if !reflect.DeepEqual(r.BindVars, expected.BindVars) {
		t.Errorf("expected bind vars %#v but got %#v", r.BindVars, expected.BindVars)
	}
}

func runTable(t *testing.T, table []testTableItem) {
	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			expectResult(t, tt.matcher, tt.input, tt.result, tt.err)
		})
	}
}

func TestLiteral(t *testing.T) {
	table := []testTableItem{
		{
			name:    "must match literal length",
			matcher: Literal("hello"),
			input:   "hello",
			result: &MatchResult{
				Match: "hello",
			},
		},
		{
			name:    "must match partial string",
			matcher: Literal("hell"),
			input:   "hello world",
			result: &MatchResult{
				Match: "hell",
				Rest:  "o world",
			},
		},
	}

	runTable(t, table)
}

func TestRepeat(t *testing.T) {
	table := []testTableItem{
		{
			name:    "must repeat consume character",
			matcher: Repeat(SingleChar('j'), false),
			input:   "jjjaansen",
			result: &MatchResult{
				Match: "jjj",
				Rest:  "aansen",
			},
		},
		{
			name:    "must allow empty matches when allowEmpty is true",
			matcher: Repeat(SingleChar('h'), true),
			input:   "w",
			result: &MatchResult{
				Match: "",
				Rest:  "w",
			},
		},
	}

	runTable(t, table)
}

func TestNegativeAssertion(t *testing.T) {
	table := []testTableItem{
		{
			name:    "must not consume any characters",
			matcher: NegativeAssert(Literal("h")),
			input:   "world",
			result: &MatchResult{
				Match: "",
				Rest:  "world",
			},
		},
	}

	runTable(t, table)
}

func TestChar(t *testing.T) {
	table := []testTableItem{
		{
			name: "must match single character from A-Z",
			matcher: Char(
				[]CharRange{
					{Start: 'A', End: 'Z'},
				},
				false,
			),
			input: "W",
			result: &MatchResult{
				Match: "W",
			},
		},
		{
			name: "must not match a-zA-Z",
			matcher: Char(
				[]CharRange{
					{Start: 'a', End: 'z'},
					{Start: 'A', End: 'Z'},
				},
				true,
			),
			input: "a",
			err:   ErrNoMatch,
		},
	}

	runTable(t, table)
}

func TestAlternation(t *testing.T) {
	table := []testTableItem{
		{
			name: "must match a|b|c",
			matcher: Alternation(
				CharIn('a', 'a'),
				CharIn('b', 'b'),
				CharIn('c', 'c'),
			),
			input: "czzzzz",
			result: &MatchResult{
				Match: "c",
				Rest:  "zzzzz",
			},
		},
	}

	runTable(t, table)
}
