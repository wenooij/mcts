package searchops

import (
	"math"

	"github.com/wenooij/heapordered"
	"github.com/wenooij/mcts"
)

// TreeMapper represents a function which maps a Tree to a single value E.
type TreeMapper[T mcts.Counter, R any] func(*heapordered.Tree[mcts.Node[T]]) R

// MinMax is a simple structure with Min and Max fields.
type MinMax struct {
	Min float64
	Max float64
}

// MinMax takes a scalar TreeReducer and returns a reducer which returns the min and max values.
//
// Note that MinMaxReducer returns a stateful TreeReducer which cannot be reused.
func MinMaxReducer[T mcts.Counter](f TreeMapper[T, float64]) TreeMapper[T, MinMax] {
	res := MinMax{math.Inf(+1), math.Inf(-1)}
	return func(e *heapordered.Tree[mcts.Node[T]]) MinMax {
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
