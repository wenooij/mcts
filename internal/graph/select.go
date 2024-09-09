package graph

import "github.com/wenooij/mcts"

// selectChild selects the highest priority child from the min heap.
func (g *graphInterface[T]) selectChild(s mcts.SearchInterface[T]) (hasChild, expand bool) {
	var n *mcts.EdgeList[T]
	switch len(g.ForwardPath) {
	case 0:
		n = g.RootEntry
	default:
		n = g.ForwardPath[len(g.ForwardPath)-1].Src
	}
	if len(*n) == 0 {
		return false, true
	}
	child := (*n)[0]
	if !s.Select(child.Action) {
		// Select may return false if this node is no longer legal
		// Possibly due to the outcome of chance node higher up the tree.
		// In SearchHash, Select may return false after a cycle is detected
		// or after a maximum depth is reached.
		//
		// In either case return child = nil, expand = false, then
		// backprop the score from n.
		return false, false
	}
	g.ForwardPath = append(g.ForwardPath, child)
	if child.Dst == nil {
		// Insert initial node.
		// We couldn't do this in Expand because Hash
		// expects to be called only after Select.
		h := s.Hash()
		// Dst will already be in Table if dst is a transposition.
		dst, ok := g.Table[h]
		if !ok {
			dst = &mcts.EdgeList[T]{}
			g.Table[h] = dst
			if g.InverseTable != nil {
				g.InverseTable[dst] = h
			}
		}
		child.Dst = dst
	}
	initializeScore(s, child)
	return true, false
}

// initializeScore is called when selecting a node for the first time.
//
// precondition: n must be the current node (s.Select(n.Action) has been called, or we are at the root).
func initializeScore[T mcts.Counter](s mcts.SearchInterface[T], e *mcts.Edge[T]) {
	if e.Score.Objective == nil {
		// E will be heapified on the first call to backprop.
		e.Score = s.Score()
	}
}
