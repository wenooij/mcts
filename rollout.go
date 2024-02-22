package mcts

// rollout runs simulated rollouts from the given node and returns the results.
func rollout[T Counter](s *Search[T], n *TableEntry[T]) (counters T, numRollouts float64) {
	if s.RolloutInterface != nil {
		// Call the custom Rollout implementation if available.
		return s.Rollout()
	}
	// Rollout using the default policy (using Expand).
	for {
		var ok bool
		switch actions := s.Expand(1); len(actions) {
		case 0:
			// Return the score for the terminal position.
			return s.Score().Counter, 1
		case 1:
			ok = s.Select(actions[0].Action)
		default:
			ok = s.Select(actions[s.Rand.Intn(len(actions))].Action)
		}
		if !ok {
			// Return the score for the terminal position.
			return s.Score().Counter, 1
		}
	}
}
