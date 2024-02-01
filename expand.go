package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

// expand calls Expand in the search interface to get more moves.
//
// the fringe argument is set to true during the rollout phase.
func expand(s *Search, n *heapordered.Tree[*node]) *heapordered.Tree[*node] {
	actions := s.Expand(0)

	if len(actions) == 0 {
		// Set the terminal node bit.
		n.Elem().nodeType |= nodeTerminal
		return nil
	}
	// Avoid bias from move generation order.
	s.Rand.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })

	// Clear terminal bit.
	n.Elem().nodeType &= ^nodeTerminal
	var (
		totalWeight    float64
		uniformWeight  float64
		uniformWeights = true
	)
	for i, a := range actions {
		child, _ := getOrCreateChild(s, n, a)
		w := child.Elem().weight
		if i == 0 {
			uniformWeight = w
		} else if a.Weight != uniformWeight {
			uniformWeights = false
		}
		// Sum predictor weights to later normalize.
		totalWeight += w
	}
	// Are the weights uniform?
	// Rosin (2.3) warns of worst-case regret in the uniform case.
	// Set all weights to 1/√K.
	if uniformWeights && len(n.Elem().childSet) > 1 {
		w := 1 / math.Sqrt(float64(len(n.Elem().childSet)))
		for _, child := range n.Elem().childSet {
			child.Elem().weight = w
		}
	} else {
		// Normalize predictor weight.
		for _, child := range n.Elem().childSet {
			child.Elem().weight /= totalWeight
		}
	}
	// Select an element to expand.
	return selectChild(s, n)
}
