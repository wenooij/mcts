package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type ObjectiveFunc func([]float64) float64

func backprop(frontier *heapordered.Tree[Node], rawScore Score, numRollouts, exploreFactor float64) {
	for n := frontier; n != nil; n = n.Parent() {
		e := &n.E
		e.Score.Add(rawScore)
		e.NumRollouts += numRollouts
		n.Fix()
		if children := n.Children(); len(children) > 0 {
			updatePrioritiesPUCB(children, e.NumRollouts, exploreFactor)
			children[0].Fix()
		}
	}
}

func backpropNull(frontier *heapordered.Tree[Node], exploreFactor float64) {
	for n := frontier; n != nil; n = n.Parent() {
		if children := n.Children(); len(children) > 0 {
			updatePrioritiesPUCB(children, n.E.NumRollouts, exploreFactor)
			children[0].Fix()
		}
	}
}

func updatePrioritiesPUCB(children []*heapordered.Tree[Node], numParentRollouts, exploreFactor float64) {
	exploreTerm := exploreFactor * math.Sqrt(numParentRollouts)
	for _, child := range children {
		if child.E.NumRollouts > 0 {
			// PUCB is used to compute the priority of a node as a combination of its mean score
			// and asymptotic priot exploration factor.
			//
			// 	PUCT(n) = Score(n) + Prior(n) * ExploreTerm(n).
			// 	Priority(n) = -PUCT(n).
			//
			// The next call to Fix will heapify E.
			nf := 1 / child.E.NumRollouts
			child.Priority = -child.E.rawScoreValue()*nf - child.E.PriorWeight*exploreTerm*nf
		}
	}
}
