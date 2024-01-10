package mcts

import "github.com/wenooij/heapordered"

func expand[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	e, _ := n.Elem()
	m := s.Expand(s.expandBuffer)
	if m == 0 {
		// Record the terminal state for later use in PV analysis.
		e.terminal = true
		return nil
	}
	for _, step := range s.expandBuffer[:m] {
		expandStep(s, n, step)
	}
	// Select the best child yet by MAB policy.
	return n.Min()
}

func expandStep[S Step](s *Search[S], parent *heapordered.Tree[*node[S]], step FrontierStep[S]) {
	// Handle the step.
	// Hit or miss based on whether we've seen it before.
	e, _ := parent.Elem()
	if _, created := getOrCreateChild(s, parent, step); created {
		e.Miss()
	} else {
		e.Hit()
	}
}
