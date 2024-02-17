package searchops

import (
	"math/rand"

	"github.com/wenooij/heapordered"
	"github.com/wenooij/mcts"
)

type (
	TreePredicate[T mcts.Counter] func(*heapordered.Tree[mcts.Node[T]]) bool
	TreeFilter[T mcts.Counter]    func([]*heapordered.Tree[mcts.Node[T]]) []*heapordered.Tree[mcts.Node[T]]
)

// FilterPredicate returns a filter which selects subtrees entries matching f.
func FilterPredicate[T mcts.Counter](f TreePredicate[T]) TreeFilter[T] {
	return func(input []*heapordered.Tree[mcts.Node[T]]) []*heapordered.Tree[mcts.Node[T]] {
		var res []*heapordered.Tree[mcts.Node[T]]
		for _, t := range input {
			if f(t) {
				res = append(res, t)
			}
		}
		return res
	}
}

// NodePredicateFilter returns a filter which selects Nodes matching f.
func FilterNodePredicate[T mcts.Counter](f func(mcts.Node[T]) bool) TreeFilter[T] {
	return FilterPredicate(func(t *heapordered.Tree[mcts.Node[T]]) bool { return f(t.E) })
}

func Filter[T mcts.Counter](tree *heapordered.Tree[mcts.Node[T]], filters ...TreeFilter[T]) *heapordered.Tree[mcts.Node[T]] {
	if len(filters) == 0 {
		return nil
	}
	candidates := make([]*heapordered.Tree[mcts.Node[T]], tree.Len())
	for i := 0; i < tree.Len(); i++ {
		candidates[i] = tree.At(i)
	}
	for _, f := range filters {
		switch candidates = f(candidates); len(candidates) {
		case 0:
			return nil
		case 1:
			return candidates[0]
		}
	}
	// Filters were not able to reduce to a single entry.
	return nil
}

// MaxFilter returns a filter which selects the entries maximumizing f.
func MaxFilter[T mcts.Counter](f TreeMapper[T, float64]) TreeFilter[T] {
	return maxCmpFilter[T](f, func(a, b float64) int {
		if a == b {
			return 0
		}
		if a < b {
			return -1
		}
		return 1
	})
}

// MinFilter returns a filter which selects the entries maximumizing f.
func MinFilter[T mcts.Counter](f TreeMapper[T, float64]) TreeFilter[T] {
	return maxCmpFilter[T](f, func(a, b float64) int {
		if a == b {
			return 0
		}
		if a < b {
			return 1
		}
		return -1
	})
}

func maxCmpFilter[T mcts.Counter](f TreeMapper[T, float64], cmp func(a, b float64) int) TreeFilter[T] {
	return func(input []*heapordered.Tree[mcts.Node[T]]) []*heapordered.Tree[mcts.Node[T]] {
		if len(input) == 0 {
			return nil
		}
		var (
			maxEntries = []*heapordered.Tree[mcts.Node[T]]{input[0]}
			maxValue   = f(input[0])
		)
		for _, e := range input[1:] {
			value := f(e)
			if cmp := cmp(value, maxValue); cmp == 0 {
				maxEntries = append(maxEntries, e)
			} else if cmp > 0 {
				maxEntries = maxEntries[:0]
				maxEntries = append(maxEntries, e)
				maxValue = value
			}
		}
		return maxEntries
	}
}

// MaxRolloutsFilter returns a filter which selects the entries with maximum rollouts.
func MaxRolloutsFilter[T mcts.Counter]() TreeFilter[T] {
	return MaxFilter(func(t *heapordered.Tree[mcts.Node[T]]) float64 { return float64(t.E.NumRollouts) })
}

// MaxScoreFilter returns a filter which selects the entries with the best normalized score.
func MaxScoreFilter[T mcts.Counter]() TreeFilter[T] {
	return MaxFilter(func(t *heapordered.Tree[mcts.Node[T]]) float64 { return t.E.Score.Apply() / float64(t.E.NumRollouts) })
}

// MaxRawScoreFilter picks the node with the best raw score.
func MaxRawScoreFilter[T mcts.Counter]() TreeFilter[T] {
	return MaxFilter(func(t *heapordered.Tree[mcts.Node[T]]) float64 { return t.E.Score.Apply() })
}

// MinPriorityFilter picks the node with the highest raw score.
func HighestPriorityFilter[T mcts.Counter]() TreeFilter[T] {
	return MinFilter(func(t *heapordered.Tree[mcts.Node[T]]) float64 { return t.Priority })
}

// MaxDepthFilter returns a filter which selects nodes below a maximum depth.
//
// Note that MaxDepthFilter returns a stateful Filter which cannot be reused.
func MaxDepthFilter[T mcts.Counter](maxDepth int) TreeFilter[T] {
	depth := 0
	return func(input []*heapordered.Tree[mcts.Node[T]]) []*heapordered.Tree[mcts.Node[T]] {
		defer func() { depth++ }()
		if maxDepth <= depth {
			return nil
		}
		return input
	}
}

// AnyFilter returns a filter which selects a random entry.
func AnyFilter[T mcts.Counter](r *rand.Rand) TreeFilter[T] {
	return func(input []*heapordered.Tree[mcts.Node[T]]) []*heapordered.Tree[mcts.Node[T]] {
		if len(input) == 0 {
			return nil
		}
		idx := r.Intn(len(input))
		return input[idx : idx+1]
	}
}

// FirstFilter returns a filter which selects the first subtree.
func FirstFilter[T mcts.Counter]() TreeFilter[T] {
	return func(input []*heapordered.Tree[mcts.Node[T]]) []*heapordered.Tree[mcts.Node[T]] {
		if len(input) == 0 {
			return nil
		}
		return input[0:1]
	}
}
