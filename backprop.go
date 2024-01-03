package mcts

func (b *topo[S]) Backprop(log Log, numRollouts int) {
	if numRollouts == 0 {
		return
	}
	for ; b != nil; b = b.parent {
		b.Log = b.Log.Merge(log)
		b.numRollouts += float64(numRollouts)
	}
}

func (e *topo[S]) NumParentRollouts(node *topo[S]) float64 {
	if node.parent == nil {
		return 0
	}
	return float64(node.parent.numRollouts)
}

func (e *topo[S]) Score() (float64, bool) {
	if e.Log == nil {
		return 0, false
	}
	return e.Log.Score(), true
}
