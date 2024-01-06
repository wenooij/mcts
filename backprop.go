package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

func backprop[S Step](n *heapordered.Tree[*node[S]], log Log, numRollouts int) {
	if numRollouts == 0 {
		return
	}
	for ; n != nil; n = n.Parent() {
		e, _ := n.Elem()
		e.Log = e.Log.Merge(log)
		e.numRollouts += float64(numRollouts)
		e.priority = -ucb1(e.Log.Score(), e.numRollouts, float64(numRollouts)+numParentRollouts(n), e.exploreParam)
		n.Fix()
	}
}

func numParentRollouts[S Step](n *heapordered.Tree[*node[S]]) float64 {
	parent := n.Parent()
	if parent == nil {
		return 0
	}
	e, _ := parent.Elem()
	return float64(e.numRollouts)
}

func ucb1(score, numRollouts, numParentRollouts float64, explorationParameter float64) float64 {
	if numRollouts == 0 || numParentRollouts == 0 {
		return math.Inf(+1)
	}
	explore := explorationParameter * math.Sqrt(math.Log(float64(numParentRollouts))/float64(numRollouts))
	exploit := score / float64(numRollouts)
	return explore + exploit
}
