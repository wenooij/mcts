package mcts

import "math"

func (b *topo[S]) Backprop(log Log, numRollouts int) {
	if numRollouts == 0 {
		return
	}
	for ; b != nil; b = b.parent {
		b.Log = b.Log.Merge(log)
		b.numRollouts += float64(numRollouts)
		b.priority = -ucb1(b.Log.Score(), b.numRollouts, float64(numRollouts)+b.NumParentRollouts(), b.exploreParam)
		b.children.Fix()
	}
}

func (e *topo[S]) NumParentRollouts() float64 {
	if e.parent == nil {
		return 0
	}
	return float64(e.parent.numRollouts)
}

func (e *topo[S]) Score() (float64, bool) {
	if e.Log == nil {
		return 0, false
	}
	return e.Log.Score(), true
}

func ucb1(score, numRollouts, numParentRollouts float64, explorationParameter float64) float64 {
	if numRollouts == 0 || numParentRollouts == 0 {
		return math.Inf(+1)
	}
	explore := explorationParameter * math.Sqrt(math.Log(float64(numParentRollouts))/float64(numRollouts))
	exploit := score / float64(numRollouts)
	return explore + exploit
}
