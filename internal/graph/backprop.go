package graph

import (
	"math"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/internal/heap"
	"github.com/wenooij/mcts/internal/model"
)

func (g *graphInterface[T]) backprop(counter mcts.CounterInterface[T], counters T, numRollouts, exploreFactor float64) {
	for i := len(g.ForwardPath) - 1; i >= 0; i-- {
		e := g.ForwardPath[i]
		counter.Add(&e.Score.Counter, counters)
		e.NumRollouts += numRollouts
		if i > 0 {
			prev := g.ForwardPath[i-1]
			score := e.Score.Objective(e.Score.Counter)
			exploreTerm := exploreFactor * math.Sqrt(prev.NumRollouts)
			e.Priority = -model.PUCB(score, e.NumRollouts, e.PriorWeight, exploreTerm)
			// NOTE(wes):
			// The index 0 here relies on a Select step which always
			// the first element. If we wanted to add a select temperature
			// parameter, we'd need to track the actual index of the edge.
			heap.Down(*prev.Dst, 0, len(*prev.Dst))
		}
		if len(*e.Src) > 0 {
			updatePrioritiesPUCB(*e.Src, float64(e.NumRollouts), exploreFactor)
			// updatePrioritiesPUCB is monotomic for elements n[1:].
			// We can save work by calling Down only on the first element of n.
			heap.Down(*e.Src, 0, len(*e.Src))
		}
	}
}

func updatePrioritiesPUCB[T mcts.Counter](es []*mcts.Edge[T], numParentRollouts, exploreFactor float64) {
	exploreTerm := exploreFactor * math.Sqrt(numParentRollouts)
	for i := range es {
		if es[i].NumRollouts > 0 {
			// PUCB is used to compute the priority of a node as a combination of its mean score
			// and asymptotic priot exploration factor.
			//
			// 	PUCT(n) = Score(n) + Prior(n) * ExploreTerm(n).
			// 	Priority(n) = -PUCT(n).
			//
			// The next call to Down will reheapify E.
			score := es[i].Score.Objective(es[i].Score.Counter)
			es[i].Priority = -model.PUCB(score, es[i].NumRollouts, es[i].PriorWeight, exploreTerm)
		}
	}
}
