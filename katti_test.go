package katti

import (
	"reflect"
	"testing"
)

func TestLiteral(t *testing.T) {
	if err, result := Parse(Literal("hello"), "hello"); err != nil {
		t.Fatal(err)
	} else {
		if !reflect.DeepEqual(result.Match, "hello") {
			t.Fatalf("match result match is different")
		}
	}
}

func TestRepeat(t *testing.T) {
	parser := Repeat(Literal("h"), false)
	err, r := Parse(parser, "hhhw")

	if err != nil {
		t.Errorf("%#v", err)
	}

	if !reflect.DeepEqual(r.Match, "hhh") {
		t.Fail()
	}
}

func TestRepeatAllowEmptyMatch(t *testing.T) {
	parser := Repeat(Literal("h"), true)
	err, r := Parse(parser, "w")

	if err != nil {
		t.Errorf("%#v", err)
	}

	if r.Match == "" && r.Rest == "w" {
	} else {
		t.Fail()
	}
}

func TestNegativeAssertion(t *testing.T) {
	parser := NegativeAssert(Literal("h"))
	err, _ := Parse(parser, "w")

	if err != nil {
		t.Fail()
	}
}

func TestCharClassSingleAZClass(t *testing.T) {
	charClass := Char("[-a-zA-Z]")

	if err, _ := Parse(charClass, "W"); err != nil {
		t.Fatal("wrong character found")
	}
}

func TestAlternationBasic(t *testing.T) {
	alt := Alternation(
		CharRange('a', 'a'),
		CharRange('b', 'b'),
		CharRange('c', 'c'),
	)

	if err, result := Parse(alt, "czz"); err != nil {
		t.Error(err)
	} else {
		if result.Match != "c" || result.Rest != "zz" {
			t.Fatal("expected 'c', '' but found", result)
		}
	}
}
