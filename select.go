package mcts

// selectChild selects the highest priority child from the min heap.
func selectChild[T Counter](s SearchInterface[T], table map[uint64]*TableEntry[T], trajectory *[]*Edge[T], n *TableEntry[T]) (child *Edge[T], expand bool) {
	if len(*n) == 0 {
		return nil, true
	}
	child = (*n)[0]
	if !s.Select(child.Action) {
		// Select may return false if this node is no longer legal
		// Possibly due to the outcome of chance node higher up the tree.
		// In SearchHash, Select may return false after a cycle is detected
		// or after a maximum depth is reached.
		//
		// In either case return child = nil, expand = false, then
		// backprop the score from n.
		return nil, false
	}
	if child.Dst == nil {
		// Insert initial node.
		// We couldn't do this in expand because Hash
		// expects to be called only after Select.
		h := s.Hash()
		// Dst will already be in Table if dst is a transposition.
		dst, ok := table[h]
		if !ok {
			dst = &TableEntry[T]{}
			table[h] = dst
		}
		child.Dst = dst
	}
	*trajectory = append(*trajectory, child)
	initializeScore(s, child)
	return child, false
}

// initializeScore is called when selecting a node for the first time.
//
// precondition: n must be the current node (s.Select(n.Action) has been called, or we are at the root).
func initializeScore[T Counter](s SearchInterface[T], e *Edge[T]) {
	if e.Score.Objective == nil {
		// E will be heapified on the first call to backprop.
		e.Score = s.Score()
	}
}
