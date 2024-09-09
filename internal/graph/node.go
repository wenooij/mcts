package graph

import (
	"math"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/searchops"
)

// makeNode creates a tree node element.
func makeNode[T mcts.Counter](action mcts.FrontierAction) mcts.Node[T] {
	weight := action.Weight
	if weight < 0 {
		panic("mcts.Node: Predictor weight < 0 for step: " + action.Action.String())
	}
	if weight == 0 {
		weight = 1
	}
	return mcts.Node[T]{
		// Max priority for new nodes.
		// This will be recomputed after the first attempt.
		Priority:    math.Inf(-1),
		PriorWeight: weight,
		Action:      action.Action,
	}
}

type explorer[T mcts.Counter] struct{ root *mcts.EdgeList[T] }

func newExplorerInterface[T mcts.Counter](root *mcts.EdgeList[T]) *explorer[T] {
	return &explorer[T]{root}
}

func (x *explorer[T]) explore(a mcts.Action) *explorer[T] {
	for _, e := range *x.root {
		if e.Action == a {
			return newExplorerInterface(e.Dst)
		}
	}
	return nil
}

func (x *explorer[T]) descendents(f func(mcts.Node[T]) error) error {
	for _, e := range *x.root {
		if err := f(e.Node); err != nil {
			if err == searchops.ErrStopIteration {
				return nil
			}
			return err
		}
	}
	return nil
}
