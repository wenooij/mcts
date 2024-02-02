package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

func backprop(frontier *heapordered.Tree[*node], rawScore Score, numRollouts float32) {
	for n := frontier; n != nil; n = n.Parent() {
		e := n.Elem()
		e.rawScore = e.rawScore.Add(rawScore)
		e.numRollouts += numRollouts
		updatePrioritiesPUCB(n, e)
	}
}

func backpropNull(leaf *heapordered.Tree[*node]) {
	for n := leaf; n != nil; n = n.Parent() {
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

func numParentRollouts(n *heapordered.Tree[*node]) float32 {
	parent := n.Parent()
	if parent == nil {
		return 0
	}
	return float32(parent.Elem().numRollouts)
}

// exploit returns the mean win rate factor.
//
// precondition: numRollouts > 0.
func exploit(rawScore, numRollouts float32) float32 { return rawScore / numRollouts }

// explore returns the exploration optimism factor:
// a function of the ratio of rollouts and parent rollouts.
//
// precondition: numRollouts >= 0 && numParentRollouts >= 0.
func explore(numRollouts, numParentRollouts float32) float32 {
	return float32(math.Sqrt(float64(fastLog(numParentRollouts+1) / numRollouts)))
}

// predictor returns a predictor loss factor.
//
// precondition: weight > 0.
func predictor(weight float32) float32 { return 2 / weight }

// pucb is short for predictor weighted upper confidence bound on trees (PUCB).
// It was introduced as a UCT extended with priors on actions.
//
// The computation for PUCB represents the fitness of a node for being selected
// on the next iteration of search. The priority for selection in the min-heap is
// the value of -pucb here. PUCB is numerically stable and optimized to be branch-free.
//
// precondition: numRollouts >= 0 && numParentRollouts >= 0.
// precondition: weight > 0.
func pucb(rawScore, numRollouts, numParentRollouts, weight, exploreFactor float32) float32 {
	nf := 1 / numRollouts
	exploit := rawScore * nf
	explore1 := float32(math.Sqrt(float64(fastLog(numParentRollouts+1) * nf)))
	predict := predictor(weight)
	explore2 := float32(math.Sqrt(float64(fastLog(numParentRollouts+1) / numParentRollouts)))
	pucb := exploit + exploreFactor*explore1 - predict*explore2
	return pucb
}
