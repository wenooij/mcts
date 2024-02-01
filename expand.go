package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

// expand calls Expand in the search interface to get more moves.
//
// the fringe argument is set to true during the rollout phase.
func expand[S Step](s *Search[S], n *heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	moves := s.Expand(0)
	if len(moves) == 0 {
		// Set the terminal node bit.
		n.Elem().NodeType |= NodeTerminal
		return nil
	}
	// Avoid bias from move generation order.
	s.Rand.Shuffle(len(moves), func(i, j int) { moves[i], moves[j] = moves[j], moves[i] })

	// Clear terminal bit.
	n.Elem().NodeType &= ^NodeTerminal
	var (
		totalWeight    float64
		uniformWeight  float64
		uniformWeights = true
	)
	for i, step := range moves {
		child, _ := getOrCreateChild(s, n, step)
		w := child.Elem().weight
		if i == 0 {
			uniformWeight = w
		} else if step.Weight != uniformWeight {
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
