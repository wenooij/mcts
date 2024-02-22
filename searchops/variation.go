package searchops

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/wenooij/mcts"
)

// Variation is a sequence of actions with Search statistics.
//
// The first element in the Variation may be a root node.
// It will have a nil Action as well among other differences.
// Use NodeType.Root to check or Variation.TrimRoot to trim it.
type Variation[T mcts.Counter] []*mcts.Edge[T]

// First returns the first edge other than the root.
//
// First returns nil if the Variation is empty.
func (v Variation[T]) First() *mcts.Edge[T] {
	if len(v) == 0 {
		return nil
	}
	return v[0]
}

// Last returns the Last edge for this variation.
//
// Last returns nil if the variation is empty.
func (v Variation[T]) Last() *mcts.Edge[T] {
	if len(v) == 0 {
		return nil
	}
	return v[len(v)-1]
}

func (v Variation[T]) String() (s string) {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	score := math.NaN()
	if v[0].Score.Objective != nil {
		score = v[0].Score.Apply() / v[0].NumRollouts
	}
	fmt.Fprintf(&sb, "[%f]", score)
	for _, e := range v {
		fmt.Fprintf(&sb, " %s", e.Action.String())
	}
	fmt.Fprintf(&sb, " (%d)", int64(v[0].NumRollouts))
	return sb.String()
}

// FilterV creates a variation by calling filters as neccessary at every step.
//
// Filters are chained together until only one entry remains per step.
// To guarantee a line is selected, add AnyFilter as the last element in the chain.
func FilterV[T mcts.Counter](root *mcts.TableEntry[T], filters ...EdgeFilter[T]) Variation[T] {
	var res Variation[T]
	edges := *root
	for len(edges) > 0 {
		curr := FilterEdges(edges, filters...)
		if curr == nil {
			break
		}
		res = append(res, curr)
		if curr.Dst == nil {
			break
		}
		edges = *curr.Dst
	}
	return res
}

// PV returns the principal variation for this Search.
//
// This is the line that the Search has searched deepest
// and is usually the best one.
//
// Use Stat to test arbitrary sequences.
func PV[T mcts.Counter](s *mcts.Search[T], extraFilters ...EdgeFilter[T]) Variation[T] {
	var filters []EdgeFilter[T]
	filters = append(filters, MaxRolloutsHashTreeFilter[T]())
	filters = append(filters, extraFilters...)
	filters = append(filters, FirstFilter[T]())
	return FilterV[T](s.RootEntry, filters...)
}

// AnyV returns a random variation with runs for this Search.
//
// AnyV is useful for statistical sampling of the Search tree.
func AnyV[T mcts.Counter](root *mcts.TableEntry[T], r *rand.Rand) Variation[T] {
	return FilterV(root, AnyFilter[T](r))
}

// Stat returns a sequence of Search stats for the given variation according to this Search.
//
// The returned Variation stops if the next action is not present in the Search tree.
func Stat[T mcts.Counter](root *mcts.TableEntry[T], vs ...mcts.Action) Variation[T] {
	if root == nil {
		return nil
	}
	res := make(Variation[T], 0, 1+len(vs))
	for _, s := range vs {
		child := Child(root, s)
		if child == nil {
			// No existing child.
			break
		}
		// Add the StatEntry and continue down the line.
		res = append(res, child)
		root = child.Dst
	}
	return res
}

// InsertV merges a new variation into the search tree.
//
// Actions already present in the search have their scores added.
// Node priorities are recomputed using UCT.
//
// The Search is initialized if it had not already done so.
// InsertV will call root and Select as a part of inserting the variation.
func InsertV[T mcts.Counter](s *mcts.Search[T], v Variation[T]) {
	root := s.RootEntry
	s.Root()
	defer s.Root()
	for _, stat := range v {
		child := Child(root, stat.Action)
		if child == nil {
			panic("InsertV: insertion of TableEntry not yet implemented")
		} else {
			panic("InsertV: merge of TableEntry not yet implemented")
		}
	}
}

func MergeSearch[T mcts.Counter](s *mcts.Search[T], table map[uint64]*mcts.TableEntry[T]) {
	for k, v := range table {
		if _, ok := s.Table[k]; ok {
			panic("MergeSearch: merge of TableEntry not yet implemented")
		} else {
			s.Table[k] = v
		}
	}
}
