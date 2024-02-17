package mcts

import (
	"math"
	"math/rand"
	"testing"
)

func TestBackpropFeatures(t *testing.T) {
	r := rand.New(rand.NewSource(1337))
	s := &Search[float64]{
		SearchInterface: &dummySearch{BranchFactor: 1, MaxDepth: 3, Rand: r},
		Rand:            r,
		NumEpisodes:     3,
	}
	s.Search()

	// Create PV.
	var pv []Node[float64]
	{
		node := s.Tree
		pv = append(pv, node.E)
		for i := 0; i < 3; i++ {
			node = node.At(0)
			pv = append(pv, node.E)
		}
	}
	// Check PV length.
	if len(pv) != 4 {
		t.Fatalf("TestBackpropFeatures(): expected |PV| = 4, got %d", len(pv))
	}
	// Priority at root should be untouched (-∞).
	if gotP, wantP := s.Tree.Priority, math.Inf(-1); gotP != wantP {
		t.Errorf("TestBackpropFeatures(): got root Priority = %f, want %f", gotP, wantP)
	}
	// Expected number of rollouts at each PV node.
	curr := s.Tree
	next := curr.At(0)
	if gotN, wantN := next.E.NumRollouts, float64(3); gotN != wantN {
		t.Errorf("TestBackpropFeatures(): got PV[0] NumRollouts = %f, want %f", gotN, wantN)
	}
	curr = next
	for i := 0; i < 3; i++ {
		if gotN, wantN := pv[i+1].NumRollouts, float64(3-i); gotN != wantN {
			t.Errorf("TestBackpropFeatures(): got PV[%d] NumRollouts = %f, want %f", i+1, gotN, wantN)
		}
	}
}
