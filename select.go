package mcts

import "github.com/wenooij/heapordered"

// selectChild selects the highest priority child from the min heap.
func selectChild[T Counter](s *Search[T], n *heapordered.Tree[Node[T]]) (child *heapordered.Tree[Node[T]], expand bool) {
	if child = n.Min(); child == nil {
		return nil, true
	}
	if !s.Select(child.E.Action) {
		// Select may return false if this node is no longer legal
		// Possibly due to the outcome of chance node higher up the tree.
		// In that case return child = nil, expand = false, then
		// start a rollout from n.
		return nil, false
	}
	initializeScore(s, child)
	return child, false
}

// initializeScore is called when selecting a node for the first time.
//
// precondition: n must be the current node (s.Select(n.Action) has been called, or we are at the root).
func initializeScore[T Counter](s *Search[T], n *heapordered.Tree[Node[T]]) {
	if n.E.NumRollouts == 0 {
		// E will be heapified on the first call to backprop.
		e := &n.E
		e.Score = s.Score()
	}
}
