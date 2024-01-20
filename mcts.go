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

// FrontierStep wraps an expanded step with extra parameters to apply to the subtree.
//
// Priority is used to seed the initial ordering of steps in the MAB min-priority-queue.
// Ideally, the priority value for X should be set to -E[Score(X)], but in practice
// a heuristic is used. Like ExploreFactor, it is critical that the priority be roughly
// proportional to the values returned from Log.Score, otherwise a small value such as -∞
// can be used to guarantee that expanded nodes are explored at least once.
// In larger state spaces this may be counterproductive.
// Priority only affects the initial value. The next priority is recomputed in backprop.
//
// ExploreFactor defines the exploration weighting for the node and its subtree.
// If this is 0, the parent's ExploreFactor is copied. By default, a ExploreFactor is
// applied uniformly to all nodes. It is critical that ExploreFactor be roughly
// proportional to the values returned from Log.Score
type FrontierStep[S Step] struct {
	Step          S
	Priority      float64
	ExploreFactor float64
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

	// Expand returns all steps in the current state.
	//
	// Expand is called after the selection phase to expand the frontier of a leaf node.
	// If Expand returns no steps, the node is marked as a terminal.
	Expand() []FrontierStep[S]

	// Rollout performs random rollouts from the current node and returns an event Log
	// and number of rollouts performed.
	//
	// Generally numRollouts can just be 1. numRollouts can be increased if multiple Rollouts
	// per epoch is helpful. Note that the Log score reported will be divided by this number.
	//
	// Backpropagation is skipped when numRollouts is 0.
	Rollout() (log Log, numRollouts int)
}
