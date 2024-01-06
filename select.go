package mcts

import "github.com/wenooij/heapordered"

func selectChild[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) (*heapordered.Tree[*node[S]], bool) {
	// Test the expand heuristic.
	if e, _ := n.Elem(); e.Test(s, n) {
		// Heuristics suggest there may be new moves at this node
		// and expand limits do not prohibit expanding from this depth.
		return nil, false
	}
	// Otherwise, select an existing child to maximize MAB policy.
	return n.Min(), true
}
