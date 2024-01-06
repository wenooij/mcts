package mcts

import "github.com/wenooij/heapordered"

func expand[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	e, _ := n.Elem()
	steps, terminal := s.Expand()
	if terminal {
		// Handle the terminal step.
		e.terminal = true
	}
	if len(steps) == 0 {
		return nil
	}
	for _, step := range steps {
		expandStep(s, n, step)
	}
	// Select the best child yet by MAB policy.
	return n.Min()
}

func expandStep[S Step](s *Search[S], parent *heapordered.Tree[*node[S]], step S) {
	// Handle the step.
	// Hit or miss based on whether we've seen it before.
	e, _ := parent.Elem()
	if _, created := getOrCreateChild(s, parent, step); created {
		e.Miss()
	} else {
		e.Hit()
	}
}
