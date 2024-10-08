package graph

import (
	"math/rand"
	"testing"

	"github.com/wenooij/mcts"
)

func TestSelectVisitsRootActions(t *testing.T) {
	const numRootActions = 20

	r := rand.New(rand.NewSource(1337))
	s := mcts.Search[float64]{
		SearchInterface: (&dummySearch{BranchFactor: numRootActions, MaxDepth: 1, Rand: r}).Interface(),
		Rand:            r,
		NumEpisodes:     numRootActions,
	}
	s.Search()

	root := s.RootEntry
	rootChildren := make([]*mcts.Edge[float64], 0, len(*root))
	for _, e := range *root {
		rootChildren = append(rootChildren, e)
	}
	if gotActions, wantActions := len(rootChildren), numRootActions; gotActions != wantActions {
		t.Errorf("TestSelectVisitsRootActions(): got children = %d, want %d", gotActions, wantActions)
	}
	for _, child := range rootChildren {
		if gotRollouts, wantRollouts := child.NumRollouts, float64(1); gotRollouts != wantRollouts {
			t.Errorf("TestSelectVisitsRootActions(%s): got rollouts = %f, want %f", child.Action, gotRollouts, wantRollouts)
		}
	}
}
