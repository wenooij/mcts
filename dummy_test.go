package mcts

import (
	"math/rand"
	"testing"
)

func BenchmarkSearch10(b *testing.B) {
	s := Search{SearchInterface: &dummySearch{
		MaxDepth:     100,
		BranchFactor: 10,
		Rand:         rand.New(rand.NewSource(1337))},
		InitialNodeCap: 10,
	}
	s.Init()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.searchEpisode()
	}
}

func BenchmarkSearch100(b *testing.B) {
	s := Search{SearchInterface: &dummySearch{
		MaxDepth:     100,
		BranchFactor: 100,
		Rand:         rand.New(rand.NewSource(1337))},
		InitialNodeCap: 100,
	}
	s.Init()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.searchEpisode()
	}
}

func BenchmarkSearch1000(b *testing.B) {
	s := Search{SearchInterface: &dummySearch{
		MaxDepth:     100,
		BranchFactor: 1000,
		Rand:         rand.New(rand.NewSource(1337))},
		InitialNodeCap: 1000,
	}
	s.Init()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.searchEpisode()
	}
}
