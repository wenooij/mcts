package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

func backprop[T Counter](frontier *heapordered.Tree[Node[T]], add func(T, T) T, counters T, numRollouts, exploreFactor float64) {
	for n := frontier; n != nil; n = n.Parent() {
		e := &n.E
		e.Score.Counter = add(e.Score.Counter, counters)
		e.NumRollouts += numRollouts
		if n.Parent() != nil {
			n.Fix()
		}
		if n.Len() > 0 {
			updatePrioritiesPUCB(n, float64(e.NumRollouts), exploreFactor)
			// updatePrioritiesPUCB is monotomic for elements n[1:].
			// We can save work by calling Down only on the first element of n.
			n.At(0).Down()
		}
	}
}

func updatePrioritiesPUCB[T Counter](n *heapordered.Tree[Node[T]], numParentRollouts, exploreFactor float64) {
	exploreTerm := exploreFactor * math.Sqrt(numParentRollouts)
	for i := 0; i < n.Len(); i++ {
		if child := n.At(i); child.E.NumRollouts > 0 {
			// PUCB is used to compute the priority of a node as a combination of its mean score
			// and asymptotic priot exploration factor.
			//
			// 	PUCT(n) = Score(n) + Prior(n) * ExploreTerm(n).
			// 	Priority(n) = -PUCT(n).
			//
			// The next call to Down will reheapify E.
			nf := 1 / float64(child.E.NumRollouts)
			score := child.E.Score.Objective(child.E.Score.Counter)
			child.Priority = -score*nf - child.E.PriorWeight*exploreTerm*nf
		}
	}
}
