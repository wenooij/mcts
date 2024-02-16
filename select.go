package mcts

import "github.com/wenooij/heapordered"

func selectChild(s *Search, n *heapordered.Tree[Node]) *heapordered.Tree[Node] {
	child := n.Min()
	if child == nil {
		return nil
	}
	s.Select(child.E.Action)
	initializeScore(s, child)
	return child
}

// initializeScore is called when selecting a node for the first time.
//
// precondition n must be the current node (s.Select(n.Action) has been called, or we are at the root).
func initializeScore(s *Search, n *heapordered.Tree[Node]) {
	if e := &n.E; e.Score.Objective == nil {
		// E will be heapified on the first call to backprop.
		e.Score = s.Score()
	}
}
