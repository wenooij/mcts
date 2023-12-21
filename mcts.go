// Package mcts provides interfaces for multi-agent Monte-Carlo tree search (MCTS).
package mcts

import "fmt"

// Step is an atomic change which can be applied to a search tree.
//
// The zero value for a Step is a special sentinel value.
type Step interface {
	comparable
	fmt.Stringer
}

// SearchInterface is a minimal interface to MCTS tree state.
type SearchInterface[E Step] interface {
	// Log returns a new empty log for the current node.
	Log() Log

	// Apply the Step to the current node.
	//
	// Apply is called multiple times in Search and after Search completes.
	Apply(E)

	// Root resets the current search to root.
	//
	// Root is called multiple times in Search before the selection phase
	// and after Search completes.
	Root()

	// Expand returns the next Step to explore for the current node.
	//
	// Expand is called during the selection phase before the rollout.
	// Expand may return Steps in any order but must return a Step if called.
	// An empty Step marks the node is marked as terminal (among its other options).
	Expand() E

	// Rollout performs one random rollout from the current node and returns an event Log.
	//
	// Rollout is called repeatedly in Search for the random rollouts phase.
	Rollout() Log
}
