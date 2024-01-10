package mcts

import "github.com/wenooij/heapordered"

func selectChild[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	// Test the expand heuristic.
	if e, _ := n.Elem(); e.Test(s, n) {
		// Heuristics suggest there may be new moves at this node
		// and expand limits do not prohibit expanding from this depth.
		return nil
	}
	// Otherwise, select an existing child to maximize MAB policy.
	return n.Min()
}
