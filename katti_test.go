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

	if expectedErr == nil && expected == nil {
		t.Fatalf("no assertions!")
	}

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
		expected.BindVars = []BindVar{}
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
			matcher: Repeat(Char('j'), false),
			input:   "jjjaansen",
			result: &MatchResult{
				Match: "jjj",
				Rest:  "aansen",
			},
		},
		{
			name:    "must allow empty matches when allowEmpty is true",
			matcher: Repeat(Char('h'), true),
			input:   "w",
			result: &MatchResult{
				Match: "",
				Rest:  "w",
			},
		},
	}

	runTable(t, table)
}

func TestRepeatRange(t *testing.T) {
	ws := Char(' ')
	table := []testTableItem{
		{
			name:    "min=0 max=3 | too many matches causes failure",
			matcher: RepeatRange(Char('j'), 0, 3, nil),
			input:   "jjaaansen",
			err:     ErrNoMatch,
		},
		{
			name:    "min=0 max=3 | exact max matches succeeds",
			matcher: RepeatRange(Char('j'), 0, 3, nil),
			input:   "jjjansen",
			result: &MatchResult{
				Match: "jjj",
				Rest:  "ansen",
			},
		},
		{
			name:    "min=0 max=3 | fewer than max matches succeeds",
			matcher: RepeatRange(Char('j'), 0, 3, nil),
			input:   "jansen",
			result: &MatchResult{
				Match: "j",
				Rest:  "ansen",
			},
		},
		{
			name:    "min=1 max=3 | zero matches fails",
			matcher: RepeatRange(Char('j'), 1, 3, nil),
			input:   "ansen",
			err:     ErrNoMatch,
		},
		{
			name:    "min=1 max=3 | one match succeeds",
			matcher: RepeatRange(Char('j'), 1, 3, nil),
			input:   "jansen",
			result: &MatchResult{
				Match: "j",
				Rest:  "ansen",
			},
		},
		{
			name:    "min=2 max=3 | one match fails",
			matcher: RepeatRange(Char('j'), 2, 3, nil),
			input:   "jansen",
			err:     ErrNoMatch,
		},
		{
			name:    "min=2 max=3 | two matches succeeds",
			matcher: RepeatRange(Char('j'), 2, 3, nil),
			input:   "jjansen",
			result: &MatchResult{
				Match: "jj",
				Rest:  "ansen",
			},
		},
		{
			name:    "min=2 max=2 | exact match succeeds",
			matcher: RepeatRange(Char('j'), 2, 2, nil),
			input:   "jjansen",
			result: &MatchResult{
				Match: "jj",
				Rest:  "ansen",
			},
		},
		{
			name:    "min=2 max=2 | overflow fails",
			matcher: RepeatRange(Char('j'), 2, 2, nil),
			input:   "jjjansen",
			err:     ErrNoMatch,
		},
		{
			name:    "min=0 max=0 | zero-width succeeds",
			matcher: RepeatRange(Char('j'), 0, 0, nil),
			input:   "jansen",
			result: &MatchResult{
				Match: "",
				Rest:  "jansen",
			},
		},
		{
			name:    "min=1 max=-1 | unbounded upper, enough matches",
			matcher: RepeatRange(Char('j'), 1, -1, nil),
			input:   "jjjjansen",
			result: &MatchResult{
				Match: "jjjj",
				Rest:  "ansen",
			},
		},
		{
			name:    "min=1 max=-1 | unbounded upper, zero matches fails",
			matcher: RepeatRange(Char('j'), 1, -1, nil),
			input:   "ansen",
			err:     ErrNoMatch,
		},
		{
			name:    "min=0 max=-1 | fully unbounded behaves like Repeat allowEmpty",
			matcher: RepeatRange(Char('j'), 0, -1, nil),
			input:   "jjjansen",
			result: &MatchResult{
				Match: "jjj",
				Rest:  "ansen",
			},
		},
		{
			name:    "min=0 max=-1 | must match seperator",
			matcher: RepeatRange(Char('j'), 0, -1, &ws),
			input:   "j j j",
			result: &MatchResult{
				Match: "j j j",
				Rest:  "",
			},
		},
		{
			name:    "min=0 max=-1 | must fail to match trailing space",
			matcher: RepeatRange(Char('j'), 0, -1, &ws),
			input:   "j j j ",
			err:     ErrNoMatch,
		},
		{
			name:    "min=0 max=-1 | must fail to match leading space",
			matcher: RepeatRange(Char('j'), 0, -1, &ws),
			input:   " j j j",
			err:     ErrNoMatch,
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

func TestSequence(t *testing.T) {
	table := []testTableItem{
		{
			name: "must match abc",
			matcher: Sequence(
				Char('a'),
				Char('b'),
				Char('c'),
			),
			input: "abcd",
			result: &MatchResult{
				Match: "abc",
				Rest:  "d",
			},
		},
		{
			name: "must drop non pluck matchers",
			matcher: Sequence(
				Char('a'),
				Skip(Char('b')),
				Char('c'),
			),
			input: "abcd",
			result: &MatchResult{
				Match: "b",
				Rest:  "d",
			},
		},
		{
			name: "must restore input on error",
			matcher: Sequence(
				Char('a'),
				Char('b'),
				Char('b'),
				Char('c'),
			),
			input: "abb$c",
			err:   ErrNoMatch,
			result: &MatchResult{
				Match: "",
				Rest:  "abb$c",
			},
		},
	}

	runTable(t, table)
}

func TestRef(t *testing.T) {
	var expr Matcher

	atom := Literal("a")

	list := Sequence(
		Char('('),
		Optional(Ref(&expr)),
		Char(')'),
	)

	expr = Alternation(atom, list)

	if _, err := Parse(expr, "(a)"); err != nil {
		t.Error(err)
	}
}

func TestRefCrossSingle(t *testing.T) {
	var even, odd Matcher

	number := Literal("1")

	even = Sequence(number, Optional(Ref(&odd)))

	odd = Sequence(Optional(Ref(&even)), number)

	// parse a simple alternating sequence
	input := "111" // even -> odd -> even
	_, err := Parse(even, input)
	if err != nil {
		t.Errorf("failed to parse cross-recursive input %q: %v", input, err)
	}
}
