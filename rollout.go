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
	score, ok := s.SearchInterface.(ScoreInterface)
	if !ok {
		// Search does not implement Score.
		// Skip rollout and backprop.
		return nil, 0
	}
	// Rollout using the default policy (using Expand).
	for {
		steps := s.Expand()
		if len(steps) == 0 {
			// Return the score for the terminal position.
			return score.Score(), 1
		}
		s.Select(steps[s.Rand.Intn(len(steps))].Step)
	}
}
