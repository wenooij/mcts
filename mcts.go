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

// FrontierStep wraps a step with an explicit initial priority in the MAB priority data structure.
//
// Smaller values indicate higher priorities. In small state spaces this can be -∞ (i.e. all steps should
// be tried at least once.) In larger state spaces, this can be counterproductive.
// Ideally, the priority value should be set to -E[s(X)], where E[s(X)] is the expected score for node X.
//
// Priority only affects the initial value. The next priority is recomputed in backprop.
type FrontierStep[S Step] struct {
	Step     S
	Priority float64
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

	// Select the Step in the current node.
	//
	// Select is called multiple times in Search and after Search completes.
	Select(S)

	// Expand returns more Steps to explore and a boolean nonterminal which should be set to true
	// when the current node is a nonterminal.
	//
	// Expand is called after the selection phase to expand the frontier of a leaf node.
	//
	// Expand may return a subset of the available steps at any given call (including an empty slice.).
	// By default, speculative expansion will call Expand multiple times (in the selection phase).
	// It is permitted for terminal nodes to return a nonempty slice of steps.
	Expand() (steps []FrontierStep[S], nonterminal bool)

	// Rollout performs random rollouts from the current node and returns an event Log
	// and number of rollouts performed.
	//
	// Generally numRollouts can just be 1. numRollouts can be increased if multiple Rollouts
	// per epoch is helpful. Note that the Log score reported will be divided by this number.
	//
	// Backpropagation is skipped when numRollouts is 0.
	Rollout() (log Log, numRollouts int)
}
