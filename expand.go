package mcts

import "github.com/wenooij/heapordered"

// expand calls Expand in the search interface to get more moves.
//
// the fringe argument is set to true during the rollout phase.
func expand[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) {
	moves := s.Expand()
	if len(moves) == 0 {
		// Set the terminal node bit.
		n.Elem().NodeType |= NodeTerminal
		return
	}
	// Clear terminal bit.
	n.Elem().NodeType &= ^NodeTerminal
	for _, step := range moves {
		getOrCreateChild(s, n, step)
	}
}
