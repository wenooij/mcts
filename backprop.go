package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

func backprop[S Step](n *heapordered.Tree[*node[S]], log Log, numRollouts float64) {
	if numRollouts == 0 {
		return
	}
	for ; n != nil; n = n.Parent() {
		e := n.Elem()
		e.Log = e.Log.Merge(log)
		e.numRollouts += numRollouts
		e.priority = -ucb1(e.Log.Score(), e.numRollouts, numRollouts+numParentRollouts(n), e.exploreFactor)
		n.Fix()
	}
}

func backpropNull[S Step](n *heapordered.Tree[*node[S]]) {
	for ; n != nil; n = n.Parent() {
		e := n.Elem()
		e.priority = -ucb1(e.Log.Score(), e.numRollouts, numParentRollouts(n), e.exploreFactor)
		n.Fix()
	}
}

func numParentRollouts[S Step](n *heapordered.Tree[*node[S]]) float64 {
	parent := n.Parent()
	if parent == nil {
		return 0
	}
	return float64(parent.Elem().numRollouts)
}

func ucb1(score, numRollouts, numParentRollouts float64, exploreFactor float64) float64 {
	if numRollouts == 0 || numParentRollouts == 0 {
		return math.Inf(+1)
	}
	explore := exploreFactor * math.Sqrt(math.Log(float64(numParentRollouts))/float64(numRollouts))
	exploit := score / float64(numRollouts)
	return explore + exploit
}
