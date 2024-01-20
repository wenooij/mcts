package mcts

import "github.com/wenooij/heapordered"

func expand[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	e, _ := n.Elem()
	moves := s.Expand()
	if len(moves) == 0 {
		// Record the terminal state.
		e.terminal = true
		return nil
	}
	for _, step := range moves {
		getOrCreateChild(s, n, step)
	}
	// Select the highest priority child by MAB policy.
	// (May or may not be a newly expanded one).
	return n.Min()
}
