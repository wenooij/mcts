// Package mcts provides interfaces for multi-agent Monte-Carlo tree search (MCTS).
package mcts

import "plugin"

type Step = string

// SearchInterface is a minimal interface to MCTS tree state.
type SearchInterface interface {
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
