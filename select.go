package mcts

func (t *topo[S]) Select(s *Search[S]) (*topo[S], bool) {
	// Test the expand heuristic.
	if t.Test(s) {
		// Heuristics suggest there may be new moves at this node
		// and expand limits do not prohibit expanding from this depth.
		return nil, false
	}
	// Otherwise, select an existing child to maximize MAB policy.
	return t.children.Min().Elem()
}
