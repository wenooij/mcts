package mcts

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/wenooij/heapordered"
)

type StatEntry[E Action] struct {
	NodeType          NodeType
	Action            E
	Score             float64
	RawScore          Score
	NumRollouts       float64
	NumParentRollouts float64
	PredictorWeight   float64
	ExploreFactor     float64
	Priority          float64
	ExploitTerm       float64
	ExploreTerm       float64
	PredictorTerm     float64
	NumChildren       int
}

func makeStatEntry[E Action](n *heapordered.Tree[*node[E]]) StatEntry[E] {
	e := n.Elem()
	return StatEntry[E]{
		NodeType:          e.nodeType,
		Action:            e.Action,
		Score:             e.NormScore(),
		RawScore:          e.rawScore,
		NumRollouts:       e.numRollouts,
		NumParentRollouts: numParentRollouts(n),
		Priority:          e.priority,
		PredictorWeight:   e.weight,
		ExploreFactor:     e.exploreFactor,
		ExploitTerm:       exploit(e.RawScore(), e.numRollouts),
		ExploreTerm:       explore(e.numRollouts, numParentRollouts(n)),
		PredictorTerm:     predictor(e.weight),
		NumChildren:       n.Len(),
	}
}

func (e StatEntry[E]) appendString(sb *strings.Builder) {
	fmt.Fprintf(sb, "[%-4.3f] %-6s (", e.Score, e.Action)
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

func (e StatEntry[E]) String() string {
	var sb strings.Builder
	e.appendString(&sb)
	return sb.String()
}

// Filter returns matching stat entries from the input.
type Filter[E Action] func([]StatEntry[E]) []StatEntry[E]

// PredicateFilter returns a filter which selects entries matching f.
func PredicateFilter[E Action](f func(StatEntry[E]) bool) Filter[E] {
	return func(input []StatEntry[E]) []StatEntry[E] {
		var res []StatEntry[E]
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
func (r Search[E]) FilterV(filters ...Filter[E]) Variation[E] {
	var res Variation[E]
	node := r.root
	res = append(res, makeStatEntry(node))
	for node != nil {
		e := filterStatNode(node, filters...)
		if e == nil {
			break
		}
		res = append(res, *e)
		node = getChild(node, e.Action)
	}
	return res
}

func filterStatNode[E Action](node *heapordered.Tree[*node[E]], filters ...Filter[E]) *StatEntry[E] {
	stat := make([]StatEntry[E], 0, node.Len())
	for _, n := range node.Elem().childSet {
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

// Best returns one of the best actions for this Search root or nil.
func (r Search[E]) Best() *StatEntry[E] {
	if r.root == nil {
		return nil
	}
	return filterStatNode(r.root, MaxRolloutsFilter[E](), AnyFilter[E](r.Rand))
}

// MaxFilter returns a filter which selects the entries maximumizing f.
func MaxFilter[E Action](f func(e StatEntry[E]) float64) Filter[E] {
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
func MinFilter[E Action](f func(e StatEntry[E]) float64) Filter[E] {
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

func maxCmpFilter[E Action](f func(e StatEntry[E]) float64, cmp func(float64, float64) int) Filter[E] {
	return func(input []StatEntry[E]) []StatEntry[E] {
		var (
			maxEntries []StatEntry[E]
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
func MaxRolloutsFilter[E Action]() Filter[E] {
	return MaxFilter[E](func(e StatEntry[E]) float64 { return e.NumRollouts })
}

// MaxScoreFilter returns a filter which selects the entries with the best normalized score.
func MaxScoreFilter[E Action]() Filter[E] {
	return MaxFilter[E](func(e StatEntry[E]) float64 { return e.Score })
}

// MaxRawScoreFilter picks the node with the best raw score.
func MaxRawScoreFilter[E Action]() Filter[E] {
	return MaxFilter[E](func(e StatEntry[E]) float64 { return e.RawScore.Score() })
}

// MinPriorityFilter picks the node with the highest raw score.
func HighestPriorityFilter[E Action]() Filter[E] {
	return MinFilter[E](func(e StatEntry[E]) float64 { return e.Priority })
}

// AnyFilter returns a filter which selects a random entry.
func AnyFilter[E Action](r *rand.Rand) Filter[E] {
	return func(input []StatEntry[E]) []StatEntry[E] {
		if len(input) == 0 {
			return nil
		}
		return []StatEntry[E]{input[r.Intn(len(input))]}
	}
}

// Subtree returns a Search for the subtree defined by the Variation v.
//
// If not all of v is present, Subtree returns nil.
func (s Search[E]) Subtree(actions ...E) *Search[E] {
	n := s.root
	if n == nil {
		return nil
	}
	for _, step := range actions {
		child := getChild(n, step)
		if child == nil {
			// Stop here.
			// The node is not present in the subtree.
			return nil
		}
		n = child
	}
	s.root = n
	return &s
}

// Reducer is a function which transforms the entry to a element of type T.
type Reducer[E Action, T any] func(e StatEntry[E]) T

// Reduce the subtree by calling r and return the final result.
func Reduce[E Action, T any](s Search[E], r Reducer[E, T]) (res T) { return reduceNode(s.root, r) }

func ReduceV[E Action, T any](s Search[E], r Reducer[E, T], v Variation[E]) (n int, res T) {
	node := s.root
	for i, e := range v {
		if node == nil {
			return i, res
		}
		res = r(makeStatEntry(node))
		node = getChild(node, e.Action)
	}
	return len(v), res
}

func reduceNode[E Action, T any](root *heapordered.Tree[*node[E]], r Reducer[E, T]) (res T) {
	stat := makeStatEntry(root)
	res = r(stat)
	for _, e := range root.Elem().childSet {
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
func MinMaxReducer[E Action](f func(StatEntry[E]) float64) Reducer[E, MinMax] {
	res := MinMax{math.Inf(+1), math.Inf(-1)}
	return func(e StatEntry[E]) MinMax {
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
