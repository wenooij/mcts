package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

func backprop[S Step](leaf *heapordered.Tree[*node[S]], rawScore Score, numRollouts float64) {
	for n := leaf; n != nil; n = n.Parent() {
		e := n.Elem()
		e.rawScore = e.rawScore.Add(rawScore)
		e.numRollouts += numRollouts
		for _, child := range e.childSet {
			childElem := child.Elem()
			childElem.priority = -ucb1(childElem.RawScore(), childElem.numRollouts, e.numRollouts, childElem.exploreFactor)
		}
		n.Init()
	}
}

func backpropNull[S Step](leaf *heapordered.Tree[*node[S]]) {
	for n := leaf; n != nil; n = n.Parent() {
		e := n.Elem()
		for _, child := range e.childSet {
			childElem := child.Elem()
			childElem.priority = -ucb1(childElem.RawScore(), childElem.numRollouts, e.numRollouts, childElem.exploreFactor)
		}
		n.Init()
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
	if numRollouts <= 0 || numParentRollouts <= 0 {
		return math.Inf(+1)
	}
	explore := exploreFactor * math.Sqrt(math.Log(numParentRollouts)/numRollouts)
	exploit := score / numRollouts
	return explore + exploit
}
