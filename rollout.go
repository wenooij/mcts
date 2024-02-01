package mcts

import (
	"github.com/wenooij/heapordered"
)

// rollout runs simulated rollouts from the given node and returns the results.
func rollout[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) (rawScore Score, numRollouts int) {
	if rollout, ok := s.SearchInterface.(RolloutInterface); ok {
		// Call the custom Rollout implementation.
		return rollout.Rollout()
	}
	// Rollout using the default policy (using Expand).
	for {
		switch steps := s.Expand(1); len(steps) {
		case 0:
			// Return the score for the terminal position.
			return s.Score(), 1
		case 1:
			s.Select(steps[0].Step)
		default:
			s.Select(steps[s.Rand.Intn(len(steps))].Step)
		}
	}
}
