package katti_test

import (
	"math/rand"
	"strings"
	"testing"

	. "jnsn.in/katti"
)

func randomString(n int) string {
	var a strings.Builder
	for range n {
		a.WriteRune(rune(rand.Intn(1114111)))
	}
	return a.String()
}

func benchChar(expr string, input string, b *testing.B) {
	for b.Loop() {
		char := Char(expr)

		_, err := Parse(char, input)

		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkCharSimple(b *testing.B) {
	benchChar("[a-z]", "c", b)
}

func BenchmarkCharSimpleInverse(b *testing.B) {
	benchChar("[^a-z]", "A", b)
}

func BenchmarkCharComplex1(b *testing.B) {
	benchChar("[a-zROADS0-8]", "7", b)
}

func BenchmarkCharComplex2(b *testing.B) {
	benchChar("[a-zROADS0-8\\]\\[]", "[", b)
}

func BenchmarkCharComplex2Inverse(b *testing.B) {
	benchChar("[^a-zROADS0-8\\]\\[]", "@", b)
}
