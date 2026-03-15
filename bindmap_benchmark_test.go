package katti

import (
	"math/rand"
	"testing"
	"time"
)

func randomChar() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	return string(chars[rand.Intn(len(chars))])
}

func BenchmarkBindMap_Get(b *testing.B) {
	bm := &BindTable{}
	values := []string{"one", "two", "three"}
	k := randomChar()

	for range 100 {
		bm.appendBindMap()

		for range 10 {
			k := randomChar()
			bm.Set(k, values)
		}
	}

	b.ResetTimer()

	for b.Loop() {
		_ = bm.Get(k)
	}
}
