package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

func backprop(frontier *heapordered.Tree[*Node], rawScore Score, numRollouts, temp float64) {
	for n := frontier; n != nil; n = n.Parent() {
		e := n.Elem()
		e.rawScore = e.rawScore.Add(rawScore)
		e.numRollouts += numRollouts
		updatePrioritiesPUCB(n, e, temp)
	}
}

func backpropNull(frontier *heapordered.Tree[*Node], temp float64) {
	for n := frontier; n != nil; n = n.Parent() {
		updatePrioritiesPUCB(n, n.Elem(), temp)
	}
}

func updatePrioritiesPUCB(n *heapordered.Tree[*Node], e *Node, temp float64) {
	for _, child := range n.Children() {
		childElem := child.Elem()
		childElem.numParentRollouts = e.numRollouts
		childElem.priority = -childElem.PUCB(temp)
	}
	n.Init()
}

// ExploitTerm returns the mean win rate factor.
//
// precondition: numRollouts > 0.
func (n *Node) ExploitTerm() float64 { return n.rawScoreValue() / n.numRollouts }

// ExploreTerm returns the exploration 'optimism' term.
// A function of ratio of exploration of the node relative to the parent Node's.
//
// precondition: numParentRollouts >= 0.
func (n *Node) ExploreTerm() float64 {
	return math.Sqrt(float64(fastLog(float32(n.numParentRollouts+1))) / n.numRollouts)
}

// PredictTerm returns a predictor loss term for PUCT.
//
// precondition: Weight > 0.
func (n *Node) PredictTerm() float64 { return 2 / n.predictWeight }

// PredictTempTerm returns the temperature applied to the predictor term in PUCT.
func (n *Node) PredictTempTerm() float64 {
	return math.Sqrt(float64(fastLog(float32(n.numParentRollouts+9))) / n.numParentRollouts)
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
func (n *Node) PUCB(exploreTemp float64) float64 {
	nf := 1 / n.numRollouts
	exploit := n.rawScoreValue() * nf
	explore := math.Sqrt(float64(fastLog(float32(n.numParentRollouts+1))) * nf)
	predict := n.PredictTerm()
	perdictTemp := n.PredictTempTerm()
	pucb := exploit + exploreTemp*float64(n.exploreFactor)*explore - float64(predict)*perdictTemp
	return pucb
}
