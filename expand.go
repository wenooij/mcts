package mcts

func (e *topo[S]) Expand(s *Search[S]) *topo[S] {
	steps, terminal := s.Expand()
	if terminal {
		// Handle the terminal step.
		e.terminal = true
	}
	if len(steps) == 0 {
		return nil
	}
	for _, step := range steps {
		e.expandStep(s, step)
	}
	// Select the best child yet by MAB policy.
	return e.childByPolicy(s)
}

func (t *topo[S]) expandStep(s *Search[S], step S) {
	// Handle the step.
	// Hit or miss based on whether we've seen it before.
	if _, created := t.newChild(s, step); created {
		t.Miss()
	} else {
		t.Hit()
	}
}
