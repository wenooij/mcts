package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

// expand calls Expand in the search interface to get more moves.
//
// the fringe argument is set to true during the rollout phase.
func expand(s *Search, n *heapordered.Tree[Node]) *heapordered.Tree[Node] {
	actions := s.Expand(0)
	e := &n.E
	defer func() { e.nodeType |= nodeExpanded }()

	if len(actions) == 0 {
		// Set the terminal node bit.
		e.nodeType |= nodeTerminal
		return nil
	}
	// Avoid bias from generation order.
	s.Rand.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })

	// Clear terminal bit.
	e.nodeType &= ^nodeTerminal
	var (
		totalWeight    float64
		uniformWeight  float64
		uniformWeights = true
	)
	for i, a := range actions {
		child, _ := getOrCreateChild(s, n, a)
		w := child.E.predictWeight
		if i == 0 {
			uniformWeight = w
		} else if w != uniformWeight {
			uniformWeights = false
		}
		child.Grow(100)
		// Sum predictor weights to later normalize.
		totalWeight += w
	}
	// Are the weights uniform?
	// Rosin (2.3) warns of worst-case regret in the uniform case.
	// Set all weights to 1/√K.
	if uniformWeights && len(n.Children()) > 1 {
		w := 1 / math.Sqrt(float64(len(n.Children())))
		for _, child := range n.Children() {
			e := &child.E
			e.predictWeight = w
		}
	} else {
		// Normalize predictor weight.
		for _, child := range n.Children() {
			e := &child.E
			e.predictWeight /= totalWeight
		}
	}
	// Select an element to expand.
	return selectChild(s, n)
}
