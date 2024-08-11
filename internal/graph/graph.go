// Package graph provides an internal interface for the graph model topology.
package graph

import "github.com/wenooij/mcts"

func InternalInterface[T mcts.Counter]() mcts.InternalInterface[T] {
	return mcts.InternalInterface[T]{
		Backprop:    backprop[T],
		Rollout:     rollout[T],
		Expand:      expand[T],
		SelectChild: selectChild[T],
		MakeNode:    makeNode[T],
	}
}
