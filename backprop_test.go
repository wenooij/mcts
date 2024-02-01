package mcts

import (
	"math"
	"math/rand"
	"testing"
)

func TestBackpropFeatures(t *testing.T) {
	r := rand.New(rand.NewSource(1337))
	s := &Search[dummyAction]{
		SearchInterface: dummySearch{B: 1, Rand: r},
		Rand:            r,
		NumEpisodes:     3,
	}
	s.Search()

	// Check PV length.
	gotPV := s.PV()
	if len(gotPV) != 4 {
		t.Fatalf("TestBackpropFeatures(): expected |PV| = 4, got %d", len(gotPV))
	}
	// Priority at root should be untouched (-∞).
	if gotP, wantP := gotPV[0].Priority, math.Inf(-1); gotP != wantP {
		t.Errorf("TestBackpropFeatures() got root Priority = %f, want %f", gotP, wantP)
	}
	// Expected number of rollouts at each PV node.
	if gotN, wantN := gotPV[0].NumRollouts, 3.0; gotN != wantN {
		t.Errorf("TestBackpropFeatures() got PV[0] NumRollouts = %f, want %f", gotN, wantN)
	}
	for i := 0; i < 3; i++ {
		if gotN, wantN := gotPV[i+1].NumRollouts, float64(3-i); gotN != wantN {
			t.Errorf("TestBackpropFeatures() got PV[%d] NumRollouts = %f, want %f", i+1, gotN, wantN)
		}
	}
}
