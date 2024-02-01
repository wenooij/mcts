package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

func backprop[S Step](frontier *heapordered.Tree[*node[S]], rawScore Score, numRollouts float64) {
	for n := frontier; n != nil; n = n.Parent() {
		e := n.Elem()
		e.rawScore = addScore(e.rawScore, rawScore)
		e.numRollouts += numRollouts
		updatePrioritiesPUCB(n, e)
	}
}

func addScore(a, b Score) Score {
	if a == nil {
		return b
	}
	return a.Add(b)
}

func backpropNull[S Step](leaf *heapordered.Tree[*node[S]]) {
	for n := leaf; n != nil; n = n.Parent() {
		updatePrioritiesPUCB(n, n.Elem())
	}
}

func updatePrioritiesPUCB[S Step](n *heapordered.Tree[*node[S]], e *node[S]) {
	for _, child := range e.childSet {
		childElem := child.Elem()
		childElem.priority = -pucb(childElem.RawScore(), childElem.numRollouts, e.numRollouts, childElem.weight, e.exploreFactor)
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

// exploit returns the mean win rate factor.
//
// precondition: numRollouts > 0.
func exploit(rawScore, numRollouts float64) float64 { return rawScore / numRollouts }

// explore returns the exploration optimism factor:
// a function of the ratio of rollouts and parent rollouts.
//
// precondition: numRollouts >= 0 && numParentRollouts >= 0.
func explore(numRollouts, numParentRollouts float64) float64 {
	return math.Sqrt(float64(fastLog(float32(numParentRollouts)+1)) / numRollouts)
}

// predictor returns a predictor loss factor.
//
// precondition: weight > 0.
func predictor(weight float64) float64 { return 2 / weight }

// pucb is short for predictor weighted upper confidence bound on trees (PUCB).
// It was introduced as a UCT extended with priors on steps.
//
// The computation for PUCB represents the fitness of a node for being selected
// on the next iteration of search. The priority for selection in the min-heap is
// the value of -pucb here. PUCB is numerically stable and optimized to be branch-free.
//
// precondition: numRollouts >= 0 && numParentRollouts >= 0.
// precondition: weight > 0.
func pucb(rawScore, numRollouts, numParentRollouts, weight, exploreFactor float64) float64 {
	nf := 1 / numRollouts
	exploit := rawScore * nf
	explore1 := math.Sqrt(float64(fastLog(float32(numParentRollouts)+1)) * nf)
	predict := predictor(weight)
	explore2 := math.Sqrt(float64(fastLog(float32(numParentRollouts)+1)) / numParentRollouts)
	pucb := exploit + exploreFactor*explore1 - predict*explore2
	return pucb
}
