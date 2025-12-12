package katti_test

import (
	"testing"

	. "jnsn.in/katti"
)

func benchChar(expr string, input string, b *testing.B) {
	for b.Loop() {
		char := Char(expr)
		_, err := Parse(char, input)

		if err != nil {
			b.Fatalf("%#v", err)
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

func benchChar2(groups []Group, inverse bool, input string, b *testing.B) {
	for b.Loop() {
		char := Char2(groups, inverse)
		_, err := Parse(char, input)

		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkChar2Simple(b *testing.B) {
	benchChar2([]Group{
		{Start: 'a', End: 'z'},
	}, false, "c", b)
}

func BenchmarkChar2SimpleInverse(b *testing.B) {
	benchChar2([]Group{
		{Start: 'a', End: 'z'},
	}, true, "A", b)
}

func BenchmarkChar2Complex1(b *testing.B) {
	benchChar2([]Group{
		{Start: 'a', End: 'z'},
		{Start: 'R', End: 'R'},
		{Start: 'O', End: 'O'},
		{Start: 'A', End: 'A'},
		{Start: 'D', End: 'D'},
		{Start: 'S', End: 'S'},
		{Start: '0', End: '8'},
	}, false, "7", b)
}

func BenchmarkChar2Complex2(b *testing.B) {
	benchChar2([]Group{
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

func BenchmarkChar2Complex2Inverse(b *testing.B) {
	benchChar2([]Group{
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
