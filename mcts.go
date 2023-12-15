// Package mcts provides interfaces for multi-agent Monte-Carlo tree search (MCTS).
package mcts

import "plugin"

// StepHash is a unique hash for a Step.
//
// It is suggested to keep Step's state at 64 or fewer bits
// as there is currently no mechanism to handle collisions.
type StepHash = uint64

// Step represents an atomic change which can be applied to a search tree.
type Step interface {
	// Hash returns a unique hash for the Step.
	// This can be precomputed if a Hash is expensive to calculate.
	Hash() StepHash

	// String returns the string representation of this Step.
	String() string
}

// SearchInterface is a minimal interface to MCTS tree state.
type SearchInterface interface {
	// Log returns a new empty log for the current node.
	Log() Log

	// Apply the Step to the current node, if possible.
	// Return true if the Step has been applied, false otherwise.
	//
	// Apply is called multiple times in Search and after Search completes.
	Apply(Step) bool

	// Root resets the current search to root.
	//
	// Root is called multiple times in Search before the selection phase
	// and after Search completes.
	Root()

	// Expand returns the next Step to explore for the current node.
	//
	// Expand is called during the select-expansion phase before the rollout.
	// Expand should return steps in a manner that guarantees all Steps are
	// visited or it samples Steps fairly (for instance, at random).
	Expand() Step

	// Rollout performs one random rollout from the current node, merges
	// the Log of events, and returns it.
	//
	// Rollout is called repeatedly in Search for the random rollouts phase.
	Rollout() Log
}

func LoadPlugin(path string) (SearchInterface, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}
	sym, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, err
	}
	newPlugin := sym.(func() SearchInterface)
	return newPlugin(), nil
}
