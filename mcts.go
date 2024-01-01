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

// Log is used to keep track of the objective function value
// as well as aggregate events of interest.
type Log interface {
	// Merge the provided event Log and return the result.
	Merge(Log) Log

	// Score returns the objective function evaluation for this Log.
	Score() float64
}

// SearchInterface is a minimal interface to MCTS tree state.
type SearchInterface[S Step] interface {
	// Log returns a new empty log for the current node.
	Log() Log

	// Root resets the current search to root.
	//
	// Root is called multiple times in Search before the selection phase
	// and after Search completes.
	Root()

	// Apply the Step to the current node.
	//
	// Apply is called multiple times in Search and after Search completes.
	Apply(S)

	// Expand returns more Steps to explore for the current node and whether
	// the node is a terminal.
	//
	// Expand is called during the selection phase before the rollout.
	//
	// By default, speculative expansion will call Expand multiple times
	// during the selection phase. This allows Expand to return a subset of the options
	// at a given time.
	//
	// terminal is only recorded when set to true.
	Expand() (steps []S, terminal bool)

	// Rollout performs random rollouts from the current node and returns an event Log
	// and number of rollouts performed.
	//
	// Generally numRollouts can just be 1. numRollouts can be increased if multiple Rollouts
	// per epoch is helpful. Note that the Log score reported will be divided by this number.
	//
	// Backpropagation is skipped when numRollouts is 0.
	Rollout() (log Log, numRollouts int)
}
