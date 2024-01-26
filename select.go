package mcts

import "github.com/wenooij/heapordered"

func selectChild[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	initializeScore(s, n)
	return n.Min()
}

func initializeScore[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) {
	// Initalize score if it has not already been set via FrontierStep.
	e := n.Elem()
	if e.rawScore != nil {
		// Score has already been initialized.
		return
	}
	if score, ok := s.SearchInterface.(ScoreInterface); ok {
		// Search implements Score so we can simply call it.
		e.rawScore = score.Score()
		return
	}
	// Otherwise skip initialization.
	// We may use RolloutsInterface to initialize the node later.
}
