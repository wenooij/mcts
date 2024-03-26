package mcts

import (
	"math"
	"math/rand"
	"slices"
)

// expand calls SearchInterface.Expand to add more Action edges to the given Node.
func expand[T Counter](s SearchInterface[T], table map[uint64]*TableEntry[T], trajectory *[]*Edge[T], n *TableEntry[T], r *rand.Rand) (child *Edge[T]) {
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
		edge := &Edge[T]{Src: n, Dst: nil, Priority: math.Inf(-1), Node: makeNode[T](a)}
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
	child, _ = selectChild(s, table, trajectory, n)
	return child
}
