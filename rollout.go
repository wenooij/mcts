package mcts

import (
	"github.com/wenooij/heapordered"
)

// rollout runs simulated rollouts from the given node and returns the results.
func rollout(s *Search, n *heapordered.Tree[Node]) (rawScore Score, numRollouts float64) {
	if rollout, ok := s.SearchInterface.(RolloutInterface); ok {
		// Call the custom Rollout implementation.
		return rollout.Rollout()
	}
	// Rollout using the default policy (using Expand).
	for {
		switch actions := s.Expand(1); len(actions) {
		case 0:
			// Return the score for the terminal position.
			return s.Score(), 1
		case 1:
			s.Select(actions[0].Action)
		default:
			s.Select(actions[s.Rand.Intn(len(actions))].Action)
		}
	}
}
