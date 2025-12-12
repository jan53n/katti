package katti_test

import (
	"testing"

	. "jnsn.in/katti"
)

func benchChar(groups []CharRange, inverse bool, input string, b *testing.B) {
	for b.Loop() {
		char := Char(groups, inverse)
		_, err := Parse(char, input)

		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkCharSimple(b *testing.B) {
	benchChar([]CharRange{
		{Start: 'a', End: 'z'},
	}, false, "c", b)
}

func BenchmarkCharSimpleInverse(b *testing.B) {
	benchChar([]CharRange{
		{Start: 'a', End: 'z'},
	}, true, "A", b)
}

func BenchmarkCharComplex1(b *testing.B) {
	benchChar([]CharRange{
		{Start: 'a', End: 'z'},
		{Start: 'R', End: 'R'},
		{Start: 'O', End: 'O'},
		{Start: 'A', End: 'A'},
		{Start: 'D', End: 'D'},
		{Start: 'S', End: 'S'},
		{Start: '0', End: '8'},
	}, false, "7", b)
}

func BenchmarkCharComplex2(b *testing.B) {
	benchChar([]CharRange{
		{Start: 'a', End: 'z'},
		{Start: 'R', End: 'R'},
		{Start: 'O', End: 'O'},
		{Start: 'A', End: 'A'},
		{Start: 'D', End: 'D'},
		{Start: 'S', End: 'S'},
		{Start: '0', End: '8'},
		{Start: ']', End: ']'},
		{Start: '[', End: '['},
	}, false, "[", b)
}

func BenchmarkCharComplex2Inverse(b *testing.B) {
	benchChar([]CharRange{
		{Start: 'a', End: 'z'},
		{Start: 'R', End: 'R'},
		{Start: 'O', End: 'O'},
		{Start: 'A', End: 'A'},
		{Start: 'D', End: 'D'},
		{Start: 'S', End: 'S'},
		{Start: '0', End: '8'},
		{Start: ']', End: ']'},
		{Start: '[', End: '['},
	}, true, "@", b)
}
