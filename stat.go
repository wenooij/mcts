package mcts

import (
	"fmt"
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

	NumChildren      int
	NumExpandHits    float64
	NumExpandSamples float64
}

func makeStatEntry[S Step](n *heapordered.Tree[*node[S]]) StatEntry[S] {
	e, _ := n.Elem()
	return StatEntry[S]{
		Step:             e.Step,
		Log:              e.Log,
		RawScore:         e.Log.Score(),
		Score:            e.NormScore(),
		NumRollouts:      e.numRollouts,
		Priority:         e.priority,
		Terminal:         e.terminal,
		NumChildren:      n.Len(),
		NumExpandHits:    float64(e.hits),
		NumExpandSamples: float64(e.Samples()),
	}
}

func (e StatEntry[S]) appendString(sb *strings.Builder) {
	fmt.Fprintf(sb, "[%-4.3f] %-6s (", e.Score, e.Step)
	// Format NumRollouts.
	switch n := e.NumRollouts; {
	case n < 1000:
		fmt.Fprintf(sb, "%.0f N; ", n)
	case n < 1e6:
		fmt.Fprintf(sb, "%.2f kN; ", n/1e3)
	default:
		fmt.Fprintf(sb, "%.2f MN; ", n/1e6)
	}
	// Format expand stats.
	fmt.Fprintf(sb, "%d children; %d samples)", e.NumChildren, int(e.NumExpandSamples))
}

// PV returns the principal variation for this Search.
//
// This is the line that the Search has searched deepest
// and is usually the best one.
//
// Use Stat to test arbitrary sequences.
func (r Search[S]) PV() Variation[S] { return r.FilterV(MaxRolloutsFilter[S](), AnyFilter[S](r.Rand)) }

// AnyV returns a random variation with runs for this Search.
//
// AnyV is useful for statistical sampling of the Search tree.
func (r Search[S]) AnyV() Variation[S] { return r.FilterV(AnyFilter[S](r.Rand)) }

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

// Variation is a sequence of Steps with Search statistics.
//
// The first element in the Variation is the root mode and will have the zero value for the Step.
type Variation[S Step] []StatEntry[S]

func (v Variation[S]) Leaf() *StatEntry[S] {
	if len(v) == 0 {
		return nil
	}
	leaf := v[len(v)-1]
	return &leaf
}

func (v Variation[S]) String() string {
	var sb strings.Builder
	if len(v) == 0 {
		return "\n"
	}
	for i := 0; i < len(v); i++ {
		e := v[i]
		fmt.Fprintf(&sb, "%-2d ", i)
		e.appendString(&sb)
		fmt.Fprintln(&sb)
	}
	return sb.String()
}

// Stat returns a sequence of Search stats for the given variation according to this Search.
//
// Stat will return all a slice of StatEntries equal to one plus the length of the input vs.
// If Search did not encounter those steps yet, the NumRollouts value will be 0.
func (r Search[S]) Stat(vs ...S) Variation[S] {
	n := r.root
	res := make(Variation[S], 0, 1+len(vs))
	res = append(res, makeStatEntry(n))
	for i, s := range vs {
		child := getChild(n, s)
		if child == nil {
			// No existing child.
			// Add dummy stat entries and break.
			for _, s := range vs[i:] {
				res = append(res, StatEntry[S]{Step: s})
			}
			break
		}
		// Add the StatEntry and continue down the line.
		n = child
		res = append(res, makeStatEntry(n))
	}
	return res
}
