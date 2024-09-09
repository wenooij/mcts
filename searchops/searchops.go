// Package searchops provides search evaluation features to the primary routines in the MCTS package.
package searchops

import (
	"errors"

	"github.com/wenooij/mcts"
)

var ErrStopIteration = errors.New("stop iteration")

// ExplorerInterface provides helpers for navigating the search node topology, outside the hot path of search,
// usually for analysis, and printing PV using the searchops package.
type Explorer[T mcts.Counter] interface {
	Walk(walkFn func(mcts.Node[T]) error) error // Walk all Nodes and call walkFn until an error is returned.
	Select(mcts.Action) bool                    // Select the Node with the given Action.
	Len() int                                   // Len returns the number of Nodes.
	At(i int) mcts.Node[T]                      // At returns the Node at i or panics.
	Ptr() any                                   // Ptr to the topological Edge.
}

// Children returns a slice of all children nodes.
func Children[T mcts.Counter](e Explorer[T]) []mcts.Node[T] {
	nodes := make([]mcts.Node[T], e.Len())
	for i := 0; i < e.Len(); i++ {
		nodes[i] = e.At(i)
	}
	return nodes
}

// Child searches the immediate children of n one-by-one and returns the subtree for a
// Otherwise returns nil if a is not present.
func Child[T mcts.Counter](n *mcts.EdgeList[T], a mcts.Action) *mcts.Edge[T] {
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
func Subtree[T mcts.Counter](root *mcts.EdgeList[T], as ...mcts.Action) *mcts.EdgeList[T] {
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
