package mcts

import "github.com/wenooij/heapordered"

func selectChild[E Action](s *Search[E], n *heapordered.Tree[*node[E]]) *heapordered.Tree[*node[E]] {
	if e := n.Elem(); e.rawScore == nil {
		// Initialize Score when selecting node for the first time.
		e.rawScore = s.Score()
	}
	return n.Min()
}
