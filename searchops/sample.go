package searchops

import (
	"math/rand/v2"

	"github.com/wenooij/mcts"
)

func UniformSample[T mcts.Counter](e Explorer[T], r *rand.Rand) (mcts.Node[T], bool) {
	n := e.Len()
	if n == 0 {
		return mcts.Node[T]{}, false
	}
	i := r.IntN(n)
	return e.At(i), true
}

func WeightedSample[T mcts.Counter](e Explorer[T], r *rand.Rand) (mcts.Node[T], bool) {
	n := e.Len()
	if n == 0 {
		return mcts.Node[T]{}, false
	}
	sum := 0.0
	weights := make([]float64, n)
	for i := 0; i < n; i++ {
		node := e.At(i)
		sum += node.NumRollouts
		weights[i] = node.NumRollouts
	}
	t := sum * r.Float64()
	for i := 0; i < n; i++ {
		node := e.At(i)
		if t -= node.NumRollouts; t <= 0 {
			return node, true
		}
	}
	panic("sampling failed unexpectedly")
}
