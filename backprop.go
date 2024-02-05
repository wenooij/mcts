package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

func backprop(frontier *heapordered.Tree[*node], rawScore Score, numRollouts float64) {
	for n := frontier; n != nil; n = n.Parent() {
		e := n.Elem()
		e.rawScore = e.rawScore.Add(rawScore)
		e.numRollouts += numRollouts
		updatePrioritiesPUCB(n, e)
	}
}

func backpropNull(frontier *heapordered.Tree[*node]) {
	for n := frontier; n != nil; n = n.Parent() {
		updatePrioritiesPUCB(n, n.Elem())
	}
}

func updatePrioritiesPUCB(n *heapordered.Tree[*node], e *node) {
	for _, child := range e.childSet {
		childElem := child.Elem()
		childElem.priority = -pucb(childElem.RawScore(), childElem.numRollouts, e.numRollouts, childElem.weight, e.exploreFactor)
	}
	n.Init()
}

func numParentRollouts(n *heapordered.Tree[*node]) float64 {
	parent := n.Parent()
	if parent == nil {
		return 0
	}
	return parent.Elem().numRollouts
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
	return math.Sqrt(float64(fastLog(float32(numParentRollouts+1))) / numRollouts)
}

// predictor returns a predictor loss factor.
//
// precondition: weight > 0.
func predictor(weight float64) float64 { return 2 / weight }

// pucb is short for predictor weighted upper confidence bound on trees (PUCB).
// It was introduced as a UCT extended with priors on actions.
//
// The computation for PUCB represents the fitness of a node for being selected
// on the next iteration of search. The priority for selection in the min-heap is
// the value of -pucb here. PUCB is numerically stable and optimized to be branch-free.
//
// precondition: numRollouts >= 0 && numParentRollouts >= 0.
// precondition: weight > 0.
func pucb(rawScore, numRollouts, numParentRollouts, weight, exploreFactor float64) float64 {
	nf := 1 / numRollouts
	exploit := float64(rawScore) * nf
	explore1 := math.Sqrt(float64(fastLog(float32(numParentRollouts+1))) * nf)
	predict := predictor(weight)
	explore2 := math.Sqrt(float64(fastLog(float32(numParentRollouts+1))) / numParentRollouts)
	pucb := exploit + float64(exploreFactor)*explore1 - float64(predict)*explore2
	return pucb
}
