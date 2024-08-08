package mcts

import (
	"math/rand"
	"testing"
)

func TestBackpropFeatures(t *testing.T) {
	r := rand.New(rand.NewSource(1337))
	s := &Search[float64]{
		SearchInterface: (&dummySearch{BranchFactor: 1, MaxDepth: 3, Rand: r}).Interface(),
		Rand:            r,
		NumEpisodes:     3,
	}
	s.Search()

	// Create PV.
	var pv []*Edge[float64]
	root := s.RootEntry
	{
		node := root
		for i := 0; i < 3; i++ {
			e := (*node)[0]
			pv = append(pv, e)
			node = e.Dst
		}
	}
	// Check PV length.
	if len(pv) != 3 {
		t.Fatalf("TestBackpropFeatures(): expected |PV| = 3, got %d", len(pv))
	}
	// Expected number of rollouts at each PV node is [3, 2, 1].
	for i, e := range pv {
		if gotN, wantN := e.NumRollouts, float64(3-i); gotN != wantN {
			t.Errorf("TestBackpropFeatures(): got PV[%d] NumRollouts = %f, want %f", i+1, gotN, wantN)
		}
	}
}
