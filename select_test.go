package mcts

import (
	"math/rand"
	"testing"
)

func TestSelectVisitsRootActions(t *testing.T) {
	const numRootActions = 20

	r := rand.New(rand.NewSource(1337))
	s := Search{
		SearchInterface: &dummySearch{BranchFactor: numRootActions, MaxDepth: 1, Rand: r},
		Rand:            r,
		NumEpisodes:     numRootActions,
	}
	s.Search()

	rootActions := s.RootActions()
	if gotActions, wantActions := len(rootActions), numRootActions; gotActions != wantActions {
		t.Errorf("TestSelectVisitsRootActions(): got actions = %d, want %d", gotActions, wantActions)
	}
	for _, a := range rootActions {
		v := s.Stat(a)
		if len(v) != 2 {
			t.Errorf("TestSelectVisitsRootActions(%s): Stat did not return a Variation containing the action", a)
			continue
		}
		e := v.Last()
		if e.Action != a {
			t.Errorf("TestSelectVisitsRootActions(%s): Stat did not return a Variation with a matching action", a)
		}
		if e.RawScore == nil {
			t.Errorf("TestSelectVisitsRootActions(%s): RawScore not initialized for action", a)
		}
		if gotRollouts, wantRollouts := e.NumRollouts, float64(1); gotRollouts != wantRollouts {
			t.Errorf("TestSelectVisitsRootActions(%s): got rollouts = %f, want %f", a, gotRollouts, wantRollouts)
		}
	}
}
