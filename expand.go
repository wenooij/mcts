package mcts

import (
	"github.com/wenooij/heapordered"
)

// expand calls Expand in the search interface to get more moves.
//
// the fringe argument is set to true during the rollout phase.
func expand[T Counter](s *Search[T], n *heapordered.Tree[Node[T]]) (child *heapordered.Tree[Node[T]]) {
	actions := s.Expand(0)
	if len(actions) == 0 {
		return nil
	}
	// Avoid bias from generation order.
	s.Rand.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })

	n.Grow(len(actions))

	var totalWeight float64
	for _, a := range actions {
		child, _ := createChild(s, n, a)
		// Sum predictor weights to later normalize.
		totalWeight += child.E.PriorWeight
	}
	// Normalize predictor weight.
	if totalWeight == 0 {
		panic("expand: got totalWeight = 0")
	}
	// Normalize the weights.
	for i := 0; i < n.Len(); i++ {
		child := n.At(i)
		e := &child.E
		e.PriorWeight /= totalWeight
	}
	// Select a child element to expand.
	child, _ = selectChild(s, n)
	return child
}
