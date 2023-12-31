package mcts

import "math/rand"

type expander[S Step] struct {
	*Search[S]
	SearchInterface[S]
	*topo[S]
	*expandHeuristic
	r *rand.Rand
}

func (e *expander[S]) Init(s *Search[S], si SearchInterface[S], n *topo[S], h *expandHeuristic, r *rand.Rand) {
	e.Search = s
	e.SearchInterface = si
	e.topo = n
	e.expandHeuristic = h
	e.r = r
}

func (e *expander[S]) Expand() *topo[S] {
	step := e.SearchInterface.Expand()
	var sentinel S
	if step == sentinel {
		// Handle the terminal step.
		// Hit or miss based on whether we've seen it before.
		if !e.terminal {
			e.Miss()
			e.terminal = true
		} else {
			e.Hit()
		}
		return nil
	}
	// Handle the step.
	// Hit or miss based on whether we've seen it before.
	child, created := e.newChild(e.Search, e.SearchInterface, step, e.r)
	if created {
		e.Miss()
	} else {
		e.Hit()
	}
	return child
}
