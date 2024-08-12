package graph

import (
	"math"
	"math/rand"
	"slices"

	"github.com/wenooij/mcts"
)

// expand calls SearchInterface.Expand to add more Action edges to the given Node.
func (g *graphInterface[T]) expand(s mcts.SearchInterface[T], path *[]*mcts.Edge[T], n *mcts.EdgeList[T], r *rand.Rand) (child *mcts.Edge[T]) {
	actions := s.Expand(0)
	if len(actions) == 0 {
		return nil
	}
	// Avoid bias from generation order.
	r.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })

	*n = slices.Grow(*n, len(actions))

	var totalWeight float64
	for _, a := range actions {
		// Dst will be filled in on the next Select.
		// We call Hash after the next Select.
		edge := &mcts.Edge[T]{Src: n, Dst: nil, Priority: math.Inf(-1), Node: makeNode[T](a)}
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
	child, _ = g.selectChild(s, path, n)
	return child
}
