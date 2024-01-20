package mcts

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/wenooij/heapordered"
)

type StatEntry[S Step] struct {
	Step        S
	Log         Log
	Score       float64
	RawScore    float64
	NumRollouts float64
	Priority    float64
	Terminal    bool
	NumChildren int
}

func makeStatEntry[S Step](n *heapordered.Tree[*node[S]]) StatEntry[S] {
	e, _ := n.Elem()
	return StatEntry[S]{
		Step:        e.Step,
		Log:         e.Log,
		RawScore:    e.Log.Score(),
		Score:       e.NormScore(),
		NumRollouts: e.numRollouts,
		Priority:    e.priority,
		Terminal:    e.terminal,
		NumChildren: n.Len(),
	}
}

func (e StatEntry[S]) appendString(sb *strings.Builder) {
	fmt.Fprintf(sb, "[%-4.3f] %-6s (", e.Score, e.Step)
	switch n := e.NumRollouts; {
	case n < 1000:
		fmt.Fprintf(sb, "%.0f N", n)
	case n < 1e6:
		fmt.Fprintf(sb, "%.2f kN", n/1e3)
	default:
		fmt.Fprintf(sb, "%.2f MN", n/1e6)
	}
	sb.WriteByte(')')
}

// Filter returns matching stat entries from the input.
type Filter[S Step] func([]StatEntry[S]) []StatEntry[S]

// PredicateFilter returns a filter which selects entries matching f.
func PredicateFilter[S Step](f func(StatEntry[S]) bool) Filter[S] {
	return func(input []StatEntry[S]) []StatEntry[S] {
		var res []StatEntry[S]
		for _, e := range input {
			if f(e) {
				res = append(res, e)
			}
		}
		return res
	}
}

// FilterV creates a variation by calling filters as neccessary at every step.
//
// Filters are chained together until only one entry remains per step.
// To guarantee a line is selected, add AnyFilter as the last element in the chain.
func (r Search[S]) FilterV(filters ...Filter[S]) Variation[S] {
	var res Variation[S]
	node := r.root
	res = append(res, makeStatEntry(node))
	for node != nil {
		e := filterStatNode(node, filters...)
		if e == nil {
			break
		}
		res = append(res, *e)
		node = getChild(node, e.Step)
	}
	return res
}

func filterStatNode[S Step](node *heapordered.Tree[*node[S]], filters ...Filter[S]) *StatEntry[S] {
	stat := make([]StatEntry[S], 0, node.Len())
	e, _ := node.Elem()
	for _, n := range e.childSet {
		stat = append(stat, makeStatEntry(n))
	}
	for _, f := range filters {
		stat = f(stat)
		if len(stat) == 0 {
			// Filters eliminated all entries.
			break
		}
		if len(stat) == 1 {
			t := stat[0]
			return &t
		}
	}
	// Filters were not able to reduce to a single entry.
	return nil
}

// Best returns one of the best Steps for this Search root or nil.
func (r Search[S]) Best() *StatEntry[S] {
	if r.root == nil {
		return nil
	}
	return filterStatNode(r.root, MaxRolloutsFilter[S](), AnyFilter[S](r.Rand))
}

// MaxFilter returns a filter which selects the entries maximumizing f.
func MaxFilter[S Step](f func(e StatEntry[S]) float64) Filter[S] {
	return maxCmpFilter(f, func(a, b float64) int {
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
func MinFilter[S Step](f func(e StatEntry[S]) float64) Filter[S] {
	return maxCmpFilter(f, func(a, b float64) int {
		if a == b {
			return 0
		}
		if a < b {
			return 1
		}
		return -1
	})
}

func maxCmpFilter[S Step](f func(e StatEntry[S]) float64, cmp func(float64, float64) int) Filter[S] {
	return func(input []StatEntry[S]) []StatEntry[S] {
		var (
			maxEntries []StatEntry[S]
			maxValue   float64
		)
		for _, e := range input {
			value := f(e)
			if cmp := cmp(value, maxValue); cmp == 0 {
				maxEntries = append(maxEntries, e)
			} else if cmp > 0 {
				maxEntries = maxEntries[:0]
				maxEntries = append(maxEntries, e)
				maxValue = e.NumRollouts
			}
		}
		return maxEntries
	}
}

// MaxRolloutsFilter returns a filter which selects the entries with maximum rollouts.
func MaxRolloutsFilter[S Step]() Filter[S] {
	return MaxFilter[S](func(e StatEntry[S]) float64 { return e.NumRollouts })
}

// MaxScoreFilter returns a filter which selects the entries with the best normalized score.
func MaxScoreFilter[S Step]() Filter[S] {
	return MaxFilter[S](func(e StatEntry[S]) float64 { return e.Score })
}

// MaxRawScoreFilter picks the node with the best raw score.
func MaxRawScoreFilter[S Step]() Filter[S] {
	return MaxFilter[S](func(e StatEntry[S]) float64 { return e.RawScore })
}

// MinPriorityFilter picks the node with the highest raw score.
func HighestPriorityFilter[S Step]() Filter[S] {
	return MinFilter[S](func(e StatEntry[S]) float64 { return e.Priority })
}

// AnyFilter returns a filter which selects a random entry.
func AnyFilter[S Step](r *rand.Rand) Filter[S] {
	return func(input []StatEntry[S]) []StatEntry[S] {
		if len(input) == 0 {
			return nil
		}
		return []StatEntry[S]{input[r.Intn(len(input))]}
	}
}

// Subtree returns a Search for the subtree defined by the Variation v.
//
// If not all of v is present, the largest subvariation is selected.
func (s Search[S]) Subtree(v Variation[S]) *Search[S] {
	n := s.root
	if n == nil {
		return &s
	}
	for _, e := range v {
		child := getChild(n, e.Step)
		if child == nil {
			// Stop here.
			// The node is not present in the subtree.
			break
		}
		n = child
	}
	s.root = n
	return &s
}

// Reducer is a function which transforms the entry to a element of type T.
type Reducer[S Step, T any] func(e StatEntry[S]) T

// Reduce the subtree by calling r and return the final result.
func Reduce[S Step, T any](s Search[S], r Reducer[S, T]) (res T) { return reduceNode(s.root, r) }

func ReduceV[S Step, T any](s Search[S], r Reducer[S, T], v Variation[S]) (n int, res T) {
	node := s.root
	for i, e := range v {
		if node == nil {
			return i, res
		}
		res = r(makeStatEntry(node))
		node = getChild(node, e.Step)
	}
	return len(v), res
}

func reduceNode[S Step, T any](root *heapordered.Tree[*node[S]], r Reducer[S, T]) (res T) {
	e, _ := root.Elem()
	stat := makeStatEntry(root)
	res = r(stat)
	for _, e := range e.childSet {
		res = reduceNode(e, r)
	}
	return res
}

// MinMax is a simple structure with Min and Max fields.
type MinMax struct {
	Min float64
	Max float64
}

// MinMax returns the min and max values in the Search subtree.
func MinMaxReducer[S Step](f func(StatEntry[S]) float64) Reducer[S, MinMax] {
	res := MinMax{math.Inf(+1), math.Inf(-1)}
	return func(e StatEntry[S]) MinMax {
		v := f(e)
		if v < res.Min {
			res.Min = v
		}
		if v > res.Max {
			res.Max = v
		}
		return res
	}
}
