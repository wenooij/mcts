package searchops

import (
	"math/rand/v2"

	"github.com/wenooij/mcts"
)

// Filter visits nodes which match all given filters.
func Filter[T mcts.Counter](e Explorer[T], visitFn func(mcts.Node[T]) error, filters ...func(mcts.Node[T]) bool) error {
	return e.Walk(func(n mcts.Node[T]) error {
		for _, f := range filters {
			if !f(n) {
				return nil
			}
		}
		visitFn(n)
		return nil
	})
}

// VariationMode picks a mode to tie break when filters select multiple variations.
type LastFilter int

const (
	NoNode    LastFilter = iota // NoNode applies no selection, stopping the variation early.
	FirstNode                   // FirstNode selects the first variation.
	AnyNode                     // AnyNode chooses the variation at random.
)

func applyLastFilter[T mcts.Counter](nodes []mcts.Node[T], r *rand.Rand, lastFilter LastFilter) (mcts.Node[T], bool) {
	if len(nodes) <= 0 {
		return mcts.Node[T]{}, false
	}
	switch lastFilter {
	case FirstNode:
		return nodes[0], true
	case AnyNode:
		return nodes[r.IntN(len(nodes))], true
	default:
		return mcts.Node[T]{}, false
	}
}

// FilterVariation visits nodes which match filters in the order specified by the game graph.
//
// Filters are applied until only one node remains.
func FilterVariation[T mcts.Counter](r *rand.Rand, ex Explorer[T], visitFn func(mcts.Node[T]) error, lastFilter LastFilter, filters ...func(mcts.Node[T]) bool) error {
	for {
		var (
			nodes []mcts.Node[T]
			ok    bool
		)
		for _, filter := range filters {
			Filter(ex, func(n mcts.Node[T]) error {
				nodes = append(nodes, n)
				ok = true
				return nil
			}, filter)
		}
		if !ok || len(nodes) <= 0 {
			break
		}
		node, ok := applyLastFilter(nodes, r, lastFilter)
		if !ok {
			break
		}
		if err := visitFn(node); err != nil {
			if err == ErrStopIteration {
				return nil
			}
			return err
		}
		ex.Select(node.Action)
	}
	return nil
}
