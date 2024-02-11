package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type ObjectiveFunc func([]float64) float64

func backprop(frontier *heapordered.Tree[Node], rawScore Score, numRollouts, exploreFactor, exploreTemp float64) {
	for n := frontier; n != nil; n = n.Parent() {
		e := &n.E
		// E will be fixed via Init in the next call to updatePrioritiesPUCB.
		// Unless n is the root node where Priority is not used.
		e.rawScore.Add(rawScore)
		e.numRollouts += numRollouts
		updatePrioritiesPUCB(n, e.numRollouts, exploreFactor, exploreTemp)
	}
}

func backpropNull(frontier *heapordered.Tree[Node], exploreFactor, exploreTemp float64) {
	for n := frontier; n != nil; n = n.Parent() {
		updatePrioritiesPUCB(n, n.E.numRollouts, exploreFactor, exploreTemp)
	}
}

func updatePrioritiesPUCB(n *heapordered.Tree[Node], numParentRollouts, exploreFactor, exploreTemp float64) {
	for _, child := range n.Children() {
		childElem := &child.E
		childElem.numParentRollouts = numParentRollouts
		// The next call to Init will heapify n.
		child.Priority = -childElem.PUCB(exploreFactor * exploreTemp)
	}
	n.Init()
}

// ExploitTerm returns the mean win rate factor.
//
// precondition: numRollouts > 0.
func (n Node) ExploitTerm() float64 { return n.Score() }

// ExploreTerm returns the exploration 'optimism' term.
// A function of ratio of exploration of the node relative to the parent Node's.
//
// precondition: numParentRollouts >= 0.
func (n Node) ExploreTerm() float64 {
	return math.Sqrt(float64(fastLog(float32(n.numParentRollouts+1))) / n.numRollouts)
}

// PredictTerm returns a predictor loss term for PUCT.
//
// precondition: Weight > 0.
func (n Node) PredictTerm() float64 { return 2 / n.predictWeight }

// PredictTempTerm returns the temperature applied to the predictor term in PUCT.
func (n Node) PredictTempTerm() float64 {
	const e1 = math.E - 1
	return math.Sqrt(float64(fastLog(float32(n.numParentRollouts+e1))) / n.numParentRollouts)
}

// PUCB is short for predictor weighted upper confidence bound on trees (PUCB).
// It was introduced as a UCT extended with priors on actions.
//
// The computation for PUCB represents the fitness of a node for being selected
// on the next iteration of search. The priority for selection in the min-heap is
// the value of -pucb here. PUCB is numerically stable and optimized to be branch-free.
//
// precondition: numRollouts >= 0 && numParentRollouts >= 0.
// precondition: weight > 0.
func (n Node) PUCB(exploreFactor float64) float64 {
	nf := 1 / n.numRollouts
	exploit := n.rawScoreValue() * nf
	explore := math.Sqrt(float64(fastLog(float32(n.numParentRollouts+1))) * nf)
	predict := n.PredictTerm()
	perdictTemp := n.PredictTempTerm()
	pucb := exploit + float64(exploreFactor)*explore - float64(predict)*perdictTemp
	return pucb
}
