package graph

import (
	"github.com/wenooij/mcts"
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
		PriorWeight: weight,
		Action:      action.Action,
	}
}
