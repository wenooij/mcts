package mcts

import (
	"math"
	"math/rand"
	"testing"
)

var tmpFastLog float64

func BenchmarkFastLog(b *testing.B) {
	var res float64
	for i := 0; i < b.N; i++ {
		for j := 0; j < 256; j++ {
			r := rand.Float64()
			res += float64(fastLog(float32(r)))
		}
	}
	tmpFastLog = res
}

var tmpLog float64

func BenchmarkLog(b *testing.B) {
	var res float64
	for i := 0; i < b.N; i++ {
		for j := 0; j < 256; j++ {
			res += math.Log(rand.Float64())
		}
	}
	tmpLog = res
}
