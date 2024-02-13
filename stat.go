package mcts

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/wenooij/heapordered"
)

func (e Node) appendString(sb *strings.Builder) {
	fmt.Fprintf(sb, "[%f] %s (%d)", e.Score(), e.Action(), int64(e.NumRollouts()))
}

func (e Node) String() string {
	var sb strings.Builder
	e.appendString(&sb)
	return sb.String()
}

// Filter returns matching stat entries from the input.
type Filter func([]Node) []Node

// PredicateFilter returns a filter which selects entries matching f.
func PredicateFilter(f func(Node) bool) Filter {
	return func(input []Node) []Node {
		var res []Node
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
func (r Search) FilterV(filters ...Filter) Variation {
	var res Variation
	node := r.Tree
	res = append(res, node.E)
	for node != nil {
		e := filterStatNode(node, filters...)
		if e == nil {
			break
		}
		res = append(res, *e)
		node = getChild(node, e.action)
	}
	return res
}

func filterStatNode(node *heapordered.Tree[Node], filters ...Filter) *Node {
	stat := make([]Node, 0, node.Len())
	for _, n := range node.Children() {
		stat = append(stat, n.E)
	}
	for _, f := range filters {
		stat = f(stat)
		if len(stat) == 0 {
			// Filters eliminated all entries.
			break
		}
		if len(stat) == 1 {
			return &stat[0]
		}
	}
	// Filters were not able to reduce to a single entry.
	return nil
}

// Best returns one of the best actions for this Search root or nil.
func (r Search) Best() *Node {
	if r.Tree == nil {
		return nil
	}
	return filterStatNode(r.Tree, MaxRolloutsFilter(), AnyFilter(r.Rand))
}

// MaxFilter returns a filter which selects the entries maximumizing f.
func MaxFilter(f func(e Node) float64) Filter {
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
func MinFilter(f func(e Node) float64) Filter {
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

func maxCmpFilter(f func(e Node) float64, cmp func(a, b float64) int) Filter {
	return func(input []Node) []Node {
		if len(input) == 0 {
			return nil
		}
		var (
			maxEntries = []Node{input[0]}
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
func MaxRolloutsFilter() Filter { return MaxFilter(func(e Node) float64 { return e.numRollouts }) }

// MaxScoreFilter returns a filter which selects the entries with the best normalized score.
func MaxScoreFilter() Filter { return MaxFilter(func(e Node) float64 { return e.Score() }) }

// MaxRawScoreFilter picks the node with the best raw score.
func MaxRawScoreFilter() Filter { return MaxFilter(func(e Node) float64 { return e.rawScore.Apply() }) }

// MinPriorityFilter picks the node with the highest raw score.
func HighestPriorityFilter() Filter {
	panic("not implemented")
}

func RootActionFilter(a Action) Filter {
	panic("not implemented")
}

func MaxDepthFilter(depth int) Filter {
	panic("not implemented")
}

// AnyFilter returns a filter which selects a random entry.
func AnyFilter(r *rand.Rand) Filter {
	return func(input []Node) []Node {
		if len(input) == 0 {
			return nil
		}
		return []Node{input[r.Intn(len(input))]}
	}
}

// Subtree returns a Search for the subtree defined by the Variation v.
//
// If not all of v is present, Subtree returns nil.
func (s Search) Subtree(actions ...Action) *Search {
	n := s.Tree
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
	s.Tree = n
	return &s
}

// Reducer is a function which transforms the entry to a element of type T.
type Reducer[T any] func(e Node) T

// Reduce the subtree by calling r and return the final result.
func Reduce[T any](s Search, r Reducer[T]) (res T) { return reduceNode(s.Tree, r) }

func ReduceV[T any](s Search, r Reducer[T], v Variation) (n int, res T) {
	node := s.Tree
	for i, e := range v {
		if node == nil {
			return i, res
		}
		res = r(node.E)
		node = getChild(node, e.action)
	}
	return len(v), res
}

func reduceNode[T any](root *heapordered.Tree[Node], r Reducer[T]) (res T) {
	stat := root.E
	res = r(stat)
	for _, e := range root.Children() {
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
func MinMaxReducer(f func(Node) float64) Reducer[MinMax] {
	res := MinMax{math.Inf(+1), math.Inf(-1)}
	return func(e Node) MinMax {
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
