package mcts

import (
	"math/rand"
	"testing"

	"github.com/wenooij/heapordered"
)

func TestSelectVisitsRootActions(t *testing.T) {
	const numRootActions = 20

	r := rand.New(rand.NewSource(1337))
	s := Search[float64]{
		SearchInterface: &dummySearch{BranchFactor: numRootActions, MaxDepth: 1, Rand: r},
		Rand:            r,
		NumEpisodes:     numRootActions,
	}
	s.Search()

	rootChildren := make([]*heapordered.Tree[Node[float64]], 0, s.Tree.Len())
	for i := 0; i < s.Tree.Len(); i++ {
		rootChildren = append(rootChildren, s.Tree.At(i))
	}
	if gotActions, wantActions := len(rootChildren), numRootActions; gotActions != wantActions {
		t.Errorf("TestSelectVisitsRootActions(): got children = %d, want %d", gotActions, wantActions)
	}
	for _, child := range rootChildren {
		if gotRollouts, wantRollouts := child.E.NumRollouts, float64(1); gotRollouts != wantRollouts {
			t.Errorf("TestSelectVisitsRootActions(%s): got rollouts = %f, want %f", child.E.Action, gotRollouts, wantRollouts)
		}
	}
}
