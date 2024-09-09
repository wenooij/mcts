package graph

import (
	"math/rand"
	"slices"

	"github.com/wenooij/mcts"
)

// expand calls SearchInterface.Expand to add more Action edges to the given Node.
func (g *graphInterface[T]) expand(s mcts.SearchInterface[T], r *rand.Rand) (hasChild bool) {
	actions := s.Expand(0)
	if len(actions) == 0 {
		return false
	}
	// Avoid bias from generation order.
	r.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })

	n := g.ForwardPath[len(g.ForwardPath)-1].Src
	*n = slices.Grow(*n, len(actions))

	var totalWeight float64
	for _, a := range actions {
		// Dst will be filled in on the next Select.
		// We call Hash after the next Select.
		edge := &mcts.Edge[T]{Src: n, Dst: nil, Node: makeNode[T](a)}
		*n = append(*n, edge)
		// Sum predictor weights to later normalize.
		totalWeight += edge.PriorWeight
	}
	// Normalize predictor weight.
	if totalWeight == 0 {
		panic("expand: got totalWeight = 0")
	}
	// Normalize the weights.
	for i := range *n {
		(*n)[i].PriorWeight /= totalWeight
	}
	// Select a child element to expand.
	hasChild, _ = g.selectChild(s)
	return hasChild
}
