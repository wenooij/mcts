package mcts

import (
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBackprop(t *testing.T) {
	r := rand.New(rand.NewSource(1337))
	s := &Search[dummyStep]{SearchInterface: dummySearch{Rand: r}, Rand: r}
	s.root = newTree(s)
	leaf := s.root
	for i := 0; i < 3; i++ {
		leaf, _ = getOrCreateChild(s, leaf, FrontierStep[dummyStep]{})
	}
	rawScore, numRollouts := rollout(s, leaf)
	backprop(leaf, rawScore, float64(numRollouts))

	const score = 0.6287385421322026
	got := s.PV()
	want := Variation[dummyStep]{{
		Step:        dummyStep(0),
		Score:       score,
		RawScore:    dummyScore(score),
		NumRollouts: 1,
		Priority:    -score,
		NumChildren: 1,
	}, {
		Step:        dummyStep(0),
		Score:       score,
		RawScore:    dummyScore(score),
		NumRollouts: 1,
		Priority:    -score,
		NumChildren: 1,
	}, {
		Step:        dummyStep(0),
		Score:       score,
		RawScore:    dummyScore(score),
		NumRollouts: 1,
		Priority:    -score,
		NumChildren: 1,
	}, {
		Step:        dummyStep(0),
		Score:       score,
		RawScore:    dummyScore(score),
		NumRollouts: 1,
		Priority:    -score,
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestBackprop(): got diff (-want, +got):\n%s", diff)
	}
}
