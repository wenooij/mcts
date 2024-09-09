package searchops

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"

	"github.com/wenooij/mcts"
)

// StripRoot removes the root node if any for use with FormatVariation
// and other functions.
func StripRoot[T any](vs []mcts.Node[T]) []mcts.Node[T] {
	if len(vs) <= 0 {
		return vs
	}
	return vs[1:]
}

// FormatVariation produces a format for a series of nodes in a game sequence.
// Use StripRoot to remove the "root node" which usually has no associated Action.
//
// The first element in the Variation may be a root node.
// It will have a nil Action as well among other differences.
// Use NodeType.Root to check or Variation.TrimRoot to trim it.
func FormatVariation[T any](vs []mcts.Node[T]) (s string) {
	var sb strings.Builder
	if len(vs) > 0 && vs[0].Score.Objective != nil {
		score := vs[0].Score.Apply() / vs[0].NumRollouts
		fmt.Fprintf(&sb, "[%f]", score)
	} else {
		sb.WriteString("[???]")
	}
	for _, v := range vs {
		fmt.Fprintf(&sb, " %s", v.Action.String())
	}
	fmt.Fprintf(&sb, " (%d)", int64(vs[0].NumRollouts))
	return sb.String()
}

// PrincipalVariation returns the main variation for this Search.
func PrincipalVariation[T mcts.Counter](ex Explorer[T], r *rand.Rand, lastFilter LastFilter) []mcts.Node[T] {
	out := []mcts.Node[T]{}
	for {
		n := ex.Len()
		if n == 0 {
			return out
		}
		maxNodes, err := Reduce(
			ValueNodesMapper[T, MinMax](MinMaxMapper(RolloutsMapper[T]())),
			ValueNodes[T, MinMax]{MinMax{math.Inf(+1), math.Inf(-1)}, nil},
			ex,
			ValueNodesReducer[T, MinMax](MinMaxReducer(RolloutsMapper[T]())))
		if err != nil || len(maxNodes.Nodes) == 0 {
			return out
		}
		node, ok := applyLastFilter(maxNodes.Nodes, r, lastFilter)
		if !ok {
			return out
		}
		out = append(out, node)
		ex.Select(node)
	}
}

// RandomVariation returns a uniform random variation with runs for this Search.
//
// RandomVariation is also useful for statistical sampling of the Search tree.
func UniformRandomVariation[T mcts.Counter](ex Explorer[T], r *rand.Rand, visitFn func(n mcts.Node[T]) error) error {
	for {
		n := ex.Len()
		if n == 0 {
			return nil
		}
		i := r.IntN(n)
		node := ex.At(i)
		if err := visitFn(node); err != nil {
			if err == ErrStopIteration {
				return nil
			}
			return err
		}
		ex.Select(node)
	}
}

// WeightedRandomVariation returns a run-weighted random variation with runs for this Search.
//
// RandomVariation is also useful for statistical sampling of the Search tree.
func WeightedRandomVariation[T mcts.Counter](ex Explorer[T], r *rand.Rand, visitFn func(n mcts.Node[T]) error) error {
	for {
		n := ex.Len()
		if n == 0 {
			return nil
		}
		i := r.IntN(n)
		node := ex.At(i)
		if err := visitFn(node); err != nil {
			if err == ErrStopIteration {
				return nil
			}
			return err
		}
		ex.Select(node)
	}
}
