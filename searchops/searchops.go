// Package searchops provides search evaluation features to the primary routines in the MCTS package.
package searchops

import (
	"github.com/wenooij/heapordered"
	"github.com/wenooij/mcts"
)

// Child searches the immediate children of n one-by-one and returns the subtree for a
// Otherwise returns nil if a is not present.
func Child[T mcts.Counter](n *heapordered.Tree[mcts.Node[T]], a mcts.Action) *heapordered.Tree[mcts.Node[T]] {
	for i := 0; i < n.Len(); i++ {
		child := n.At(i)
		if child.E.Action == a {
			return child
		}
	}
	return nil
}

// Subtree returns a the subtree defined by the input actions.
//
// If not all actions are present, Subtree returns nil.
func Subtree[T mcts.Counter](root *heapordered.Tree[mcts.Node[T]], as ...mcts.Action) *heapordered.Tree[mcts.Node[T]] {
	for _, a := range as {
		child := Child(root, a)
		if child == nil {
			// The node is not present in the subtree.
			return nil
		}
		root = child
	}
	return root
}
