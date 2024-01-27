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
		updatePrioritiesPUCB(n)
	}
}

func backpropNull[S Step](leaf *heapordered.Tree[*node[S]]) {
	for n := leaf; n != nil; n = n.Parent() {
		updatePrioritiesPUCB(n)
	}
}

func updatePrioritiesPUCB[S Step](n *heapordered.Tree[*node[S]]) {
	e := n.Elem()
	totalWeight := e.totalWeight
	for _, child := range n.Elem().childSet {
		childElem := child.Elem()
		childElem.priority = -pucb(childElem.RawScore(), childElem.numRollouts, e.numRollouts, childElem.weight, totalWeight, childElem.exploreFactor)
	}
	n.Init()
}

func numParentRollouts[S Step](n *heapordered.Tree[*node[S]]) float64 {
	parent := n.Parent()
	if parent == nil {
		return 0
	}
	return float64(parent.Elem().numRollouts)
}

func exploit(score, numRollouts float64) float64 {
	if numRollouts <= 0 {
		return math.Inf(+1)
	}
	return score / numRollouts
}

func explore(numRollouts, numParentRollouts float64) float64 {
	if numRollouts <= 0 || numParentRollouts <= 1 {
		return 0
	}
	return math.Sqrt(math.Log(numParentRollouts) / numRollouts)
}

func predictor(weight, totalWeight, explore float64) float64 {
	if totalWeight <= 0 {
		return 0
	}
	predictor := 2 / (weight / totalWeight)
	if explore != 0 {
		predictor *= explore
	}
	return predictor
}

func pucb(score, numRollouts, numParentRollouts float64, weight, totalWeight float64, exploreFactor float64) float64 {
	exploit := exploit(score, numRollouts)
	explore := explore(numRollouts, numParentRollouts)
	predict := predictor(weight, totalWeight, explore)
	return exploit + exploreFactor*explore - predict
}
