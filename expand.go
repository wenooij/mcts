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
	steps, terminal := e.SearchInterface.Expand()
	if terminal {
		// Handle the terminal step.
		e.terminal = true
	}
	if len(steps) == 0 {
		return nil
	}
	for _, s := range steps {
		e.expandStep(s)
	}
	// Select the best child yet by MAB policy.
	return e.childByPolicy(e.r)
}

func (e *expander[S]) expandStep(s S) {
	// Handle the step.
	// Hit or miss based on whether we've seen it before.
	if _, created := e.newChild(e.Search, e.SearchInterface, s, e.r); created {
		e.Miss()
	} else {
		e.Hit()
	}
}
