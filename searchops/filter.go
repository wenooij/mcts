package searchops

import (
	"math/rand"

	"github.com/wenooij/mcts"
)

type (
	EdgePredicate[T mcts.Counter] func(*mcts.Edge[T]) bool
	EdgeFilter[T mcts.Counter]    func([]*mcts.Edge[T]) []*mcts.Edge[T]
)

// FilterPredicate returns a filter which selects subtrees entries matching f.
func (f EdgePredicate[T]) Filter(input []*mcts.Edge[T]) []*mcts.Edge[T] {
	var res []*mcts.Edge[T]
	for _, t := range input {
		if f(t) {
			res = append(res, t)
		}
	}
	return res
}

func FilterEdges[T mcts.Counter](candidates []*mcts.Edge[T], filters ...EdgeFilter[T]) *mcts.Edge[T] {
	if len(filters) == 0 {
		return nil
	}
	for _, f := range filters {
		if candidates = f(candidates); len(candidates) == 0 {
			return nil
		}
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	// Filters were not able to reduce to a single entry.
	return nil
}

// MaxFilter returns a filter which selects the entries maximumizing f.
func MaxFilter[T mcts.Counter](f EdgeMapper[T, float64]) EdgeFilter[T] {
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
func MinFilter[T mcts.Counter](f EdgeMapper[T, float64]) EdgeFilter[T] {
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

func maxCmpFilter[T mcts.Counter](f EdgeMapper[T, float64], cmp func(a, b float64) int) EdgeFilter[T] {
	return func(input []*mcts.Edge[T]) []*mcts.Edge[T] {
		if len(input) == 0 {
			return nil
		}
		var (
			maxEntries = []*mcts.Edge[T]{input[0]}
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
func MaxRolloutsHashTreeFilter[T mcts.Counter]() EdgeFilter[T] {
	return MaxFilter(func(e *mcts.Edge[T]) float64 { return e.NumRollouts })
}

// MaxScoreFilter returns a filter which selects the entries with the best normalized score.
func MaxScoreFilter[T mcts.Counter]() EdgeFilter[T] {
	return MaxFilter(func(e *mcts.Edge[T]) float64 { return e.Score.Apply() / float64(e.NumRollouts) })
}

// MaxRawScoreFilter picks the node with the best raw score.
func MaxRawScoreFilter[T mcts.Counter]() EdgeFilter[T] {
	return MaxFilter(func(e *mcts.Edge[T]) float64 { return e.Score.Apply() })
}

// MinPriorityFilter picks the node with the highest raw score.
func HighestPriorityFilter[T mcts.Counter]() EdgeFilter[T] {
	return MinFilter(func(e *mcts.Edge[T]) float64 { return e.Priority })
}

// MaxDepthFilter returns a filter which selects nodes below a maximum depth.
//
// Note that MaxDepthFilter returns a stateful Filter which cannot be reused.
func MaxDepthFilter[T mcts.Counter](maxDepth int) EdgeFilter[T] {
	depth := 0
	return func(input []*mcts.Edge[T]) []*mcts.Edge[T] {
		defer func() { depth++ }()
		if maxDepth <= depth {
			return nil
		}
		return input
	}
}

// AnyFilter returns a filter which selects a random entry.
func AnyFilter[T mcts.Counter](r *rand.Rand) EdgeFilter[T] {
	return func(input []*mcts.Edge[T]) []*mcts.Edge[T] {
		if len(input) == 0 {
			return nil
		}
		idx := r.Intn(len(input))
		return input[idx : idx+1]
	}
}

// FirstFilter returns a filter which selects the first subtree.
func FirstFilter[T mcts.Counter]() EdgeFilter[T] {
	return func(input []*mcts.Edge[T]) []*mcts.Edge[T] {
		if len(input) == 0 {
			return nil
		}
		return input[0:1]
	}
}

func HasObjective[T mcts.Counter]() EdgeFilter[T] {
	return func(es []*mcts.Edge[T]) []*mcts.Edge[T] {
		var res []*mcts.Edge[T]
		for _, e := range es {
			if e.Score.Objective != nil {
				res = append(res, e)
			}
		}
		return res
	}
}

// VisitedFilter returns an EdgeFilter which filters out visited entries.
func VisitedFilter[T mcts.Counter](s *mcts.Search[T]) EdgeFilter[T] {
	visited := make(map[*mcts.TableEntry[T]]struct{})
	root := s.RootEntry
	visited[root] = struct{}{}
	return func(es []*mcts.Edge[T]) []*mcts.Edge[T] {
		var res []*mcts.Edge[T]
		for _, e := range es {
			if _, ok := visited[e.Dst]; !ok {
				res = append(res, e)
			}
		}
		if len(res) == 1 {
			visited[res[0].Dst] = struct{}{}
			return res
		}
		return res
	}
}
