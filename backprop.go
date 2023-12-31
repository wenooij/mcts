package mcts

type backprop[S Step] struct {
	parent      *backprop[S]
	Log         Log
	numRollouts float64
}

func (b *backprop[S]) Init(frontier *topo[S], log Log) {
	if frontier.parent != nil {
		b.parent = &frontier.parent.backprop
	}
	b.Log = log
}

func (b *backprop[S]) Backprop(log Log, numRollouts int) {
	for ; b != nil; b = b.parent {
		b.Log = b.Log.Merge(log)
		b.numRollouts += float64(numRollouts)
	}
}

func (e backprop[S]) NumParentRollouts() float64 {
	if e.parent == nil {
		return 0
	}
	return float64(e.parent.numRollouts)
}

func (e backprop[S]) Score() (float64, bool) {
	if e.Log == nil {
		return 0, false
	}
	return e.Log.Score(), true
}
