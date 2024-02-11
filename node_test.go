package mcts

import (
	"math/rand"
	"testing"
)

func TestNodeTypes(t *testing.T) {
	r := rand.New(rand.NewSource(1337))
	s := Search{
		SearchInterface: &dummySearch{MaxDepth: 2, BranchFactor: 2, Rand: r},
		Rand:            r,
		NumEpisodes:     100,
	}
	s.Search()
	for i := 0; i < 100; i++ {
		anyV := s.AnyV()
		for i, e := range anyV {
			switch i {
			case 0:
				if !e.Root() {
					t.Errorf("TestNodeTypes(): expected root node %d of %q", i, anyV)
				}
			case 2:
				if e.Expanded() && !e.Terminal() {
					t.Errorf("TestNodeTypes(): expected terminal node %d of %q", i, anyV)
				}
			default:
				var want nodeType
				if e.Expanded() {
					want = nodeExpanded
				}
				if e.nodeType != want {
					t.Errorf("TestNodeTypes(): internal node %d of %q has type = %v, want %v", i, anyV, e.nodeType, want)
				}
			}
		}
	}
}
