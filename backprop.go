package mcts

import "math"

func backprop[T Counter](trajectory []*Edge[T], add func(*T, T), counters T, numRollouts, exploreFactor float64) {
	for i := len(trajectory) - 1; i >= 0; i-- {
		e := trajectory[i]
		add(&e.Score.Counter, counters)
		e.NumRollouts += numRollouts
		if i > 0 {
			prev := trajectory[i-1]
			score := e.Score.Objective(e.Score.Counter)
			exploreTerm := exploreFactor * math.Sqrt(prev.NumRollouts)
			e.Priority = -pucb(score, e.NumRollouts, e.PriorWeight, exploreTerm)
			// NOTE(wes):
			// The index 0 here relies on a Select step which always
			// the first element. If we wanted to add a select temperature
			// parameter, we'd need to track the actual index of the edge.
			down(*prev.Dst, 0, len(*prev.Dst))
		}
		if len(*e.Src) > 0 {
			updatePrioritiesPUCB(*e.Src, float64(e.NumRollouts), exploreFactor)
			// updatePrioritiesPUCB is monotomic for elements n[1:].
			// We can save work by calling Down only on the first element of n.
			down(*e.Src, 0, len(*e.Src))
		}
	}
}

func updatePrioritiesPUCB[T Counter](es []*Edge[T], numParentRollouts, exploreFactor float64) {
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
			es[i].Priority = -pucb(score, es[i].NumRollouts, es[i].PriorWeight, exploreTerm)
		}
	}
}

func pucb(score, numRollouts, priorWeight, exploreTerm float64) float64 {
	nf := 1 / float64(numRollouts)
	return score*nf + priorWeight*exploreTerm*nf
}

func swap[T any](h []*Edge[T], i, j int) { h[i], h[j] = h[j], h[i] }

func down[T Counter](h []*Edge[T], i0 int, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1                                                      // Left child.
		if j2 := j1 + 1; j2 < n && h[j2].Priority < h[j1].Priority { // Less(j2, j1)
			j = j2 // = 2*i + 2  // right child
		}
		if h[j].Priority >= h[i].Priority { // !Less(j, i)
			break
		}
		swap(h, i, j)
		i = j
	}
	return i > i0
}
