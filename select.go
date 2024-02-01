package mcts

import "github.com/wenooij/heapordered"

func selectChild(s *Search, n *heapordered.Tree[*node]) *heapordered.Tree[*node] {
	if e := n.Elem(); e.rawScore == nil {
		// Initialize Score when selecting node for the first time.
		e.rawScore = s.Score()
	}
	return n.Min()
}
