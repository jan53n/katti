package main

import (
	"testing"
)

func Benchmark(b *testing.B) {
	for b.Loop() {
		if _, err := ParseSemver(); err != nil {
			b.Error(err)
		} else {
			//
		}
	}
}
