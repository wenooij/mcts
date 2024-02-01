package mcts

import "github.com/wenooij/heapordered"

func selectChild[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	if e := n.Elem(); e.rawScore == nil {
		// Initialize Score when selecting node for the first time.
		e.rawScore = s.Score()
	}
	return n.Min()
}
