package graph

import (
	"math/rand"
	"testing"

	"github.com/wenooij/mcts"
)

func BenchmarkSearch10(b *testing.B) {
	s := mcts.Search[float64]{SearchInterface: (&dummySearch{
		MaxDepth:     100,
		BranchFactor: 10,
		Rand:         rand.New(rand.NewSource(1337)),
	}).Interface()}
	b.ReportAllocs()
	s.Init()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Search()
	}
}

func BenchmarkSearch100(b *testing.B) {
	s := mcts.Search[float64]{SearchInterface: (&dummySearch{
		MaxDepth:     100,
		BranchFactor: 100,
		Rand:         rand.New(rand.NewSource(1337)),
	}).Interface()}
	b.ReportAllocs()
	s.Init()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Search()
	}
}

func BenchmarkSearch1000(b *testing.B) {
	s := mcts.Search[float64]{SearchInterface: (&dummySearch{
		MaxDepth:     100,
		BranchFactor: 1000,
		Rand:         rand.New(rand.NewSource(1337)),
	}).Interface()}
	b.ReportAllocs()
	s.Init()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Search()
	}
}
