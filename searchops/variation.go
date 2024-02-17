package searchops

import (
	"math"
	"math/rand"

	"github.com/wenooij/heapordered"
	"github.com/wenooij/mcts"
)

// FilterV creates a variation by calling filters as neccessary at every step.
//
// Filters are chained together until only one entry remains per step.
// To guarantee a line is selected, add AnyFilter as the last element in the chain.
func FilterV[T mcts.Counter](root *heapordered.Tree[mcts.Node[T]], filters ...TreeFilter[T]) mcts.Variation[T] {
	var res mcts.Variation[T]
	res = append(res, root.E)
	for root != nil {
		curr := Filter(root, filters...)
		if curr == nil {
			break
		}
		res = append(res, curr.E)
		root = curr
	}
	return res
}

// PV returns the principal variation for this Search.
//
// This is the line that the Search has searched deepest
// and is usually the best one.
//
// Use Stat to test arbitrary sequences.
func PV[T mcts.Counter](root *heapordered.Tree[mcts.Node[T]]) mcts.Variation[T] {
	return FilterV[T](root, MaxRolloutsFilter[T](), FirstFilter[T]())
}

// AnyV returns a random variation with runs for this Search.
//
// AnyV is useful for statistical sampling of the Search tree.
func AnyV[T mcts.Counter](root *heapordered.Tree[mcts.Node[T]], r *rand.Rand) mcts.Variation[T] {
	return FilterV(root, AnyFilter[T](r))
}

// Stat returns a sequence of Search stats for the given variation according to this Search.
//
// The returned Variation stops if the next action is not present in the Search tree.
func Stat[T mcts.Counter](root *heapordered.Tree[mcts.Node[T]], vs ...mcts.Action) mcts.Variation[T] {
	if root == nil {
		return nil
	}
	res := make(mcts.Variation[T], 0, 1+len(vs))
	res = append(res, root.E)
	for _, s := range vs {
		child := Child(root, s)
		if child == nil {
			// No existing child.
			break
		}
		// Add the StatEntry and continue down the line.
		res = append(res, child.E)
		root = child
	}
	return res
}

// InsertV merges a new variation into the search tree.
//
// Actions already present in the search have their scores added.
// Node priorities are recomputed using UCT.
//
// The Search is initialized if it had not already done so.
func InsertV[T mcts.Counter](root *heapordered.Tree[mcts.Node[T]], v mcts.Variation[T]) {
	// FIXME(): Insert stat at Root.
	for _, stat := range v.TrimRoot() {
		child := Child(root, stat.Action)
		var e mcts.Node[T]
		if child == nil {
			e.Action = stat.Action
			e.PriorWeight = stat.PriorWeight
			e.Score = stat.Score
			e.NumRollouts = stat.NumRollouts
			root = root.NewChild(e, math.Inf(-1))
		} else {
			// FIXME: Add stat value to node.
			child.Init()
			root = child
		}
	}
}
