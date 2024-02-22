// Package searchops provides search evaluation features to the primary routines in the MCTS package.
package searchops

import (
	"github.com/wenooij/mcts"
)

// Child searches the immediate children of n one-by-one and returns the subtree for a
// Otherwise returns nil if a is not present.
func Child[T mcts.Counter](n *mcts.TableEntry[T], a mcts.Action) *mcts.Edge[T] {
	for _, e := range *n {
		if e.Action == a {
			return e
		}
	}
	return nil
}

// Subtree returns a the subtree defined by the input actions.
//
// If not all actions are present, Subtree returns nil.
func Subtree[T mcts.Counter](root *mcts.TableEntry[T], as ...mcts.Action) *mcts.TableEntry[T] {
	for _, a := range as {
		child := Child(root, a)
		if child == nil {
			// The node is not present in the subtree.
			return nil
		}
		root = child.Dst
	}
	return root
}
