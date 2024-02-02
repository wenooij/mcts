package mcts

import "github.com/wenooij/heapordered"

func selectChild(s *Search, n *heapordered.Tree[*node]) *heapordered.Tree[*node] {
	child := n.Min()
	if child == nil {
		return nil
	}
	s.Select(child.Elem().Action)
	initializeScore(s, child)
	return child
}

// initializeScore is called when selecting a node for the first time.
//
// precondition n must be the current node (s.Select(n.Action) has been called, or we are at the root).
func initializeScore(s *Search, n *heapordered.Tree[*node]) {
	if e := n.Elem(); e.rawScore == nil {
		e.rawScore = s.Score()
	}
}
