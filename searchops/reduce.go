package searchops

import (
	"github.com/wenooij/mcts"
)

// Reduce uses the stateless reducers and
func Reduce[T mcts.Counter, R any](m func(mcts.Node[T]) R, r0 R, ex Explorer[T], reducers ...func(R, R) R) (R, error) {
	res := r0
	err := ex.Walk(func(n mcts.Node[T]) error {
		b := m(n)
		for _, r := range reducers {
			res = r(res, b)
		}
		return nil
	})
	if err == ErrStopIteration {
		err = nil
	}
	return res, err
}

// MinMaxReducer creates a stateless Reducer function which minimizes and maximizes m.
//
// See the available Mappers to use with this function.
func MinMaxReducer[T mcts.Counter](m func(mcts.Node[T]) float64) func(MinMax, MinMax) MinMax {
	return func(res, b MinMax) MinMax {
		if b.Min < res.Min {
			res.Min = b.Min
		}
		if b.Max > res.Max {
			res.Max = b.Max
		}
		return res
	}
}

func ValueNodeReducer[T mcts.Counter, E comparable](r func(E, E) E) func(res, b ValueNode[T, E]) ValueNode[T, E] {
	return func(res, b ValueNode[T, E]) ValueNode[T, E] {
		rv := r(res.Value, b.Value)
		if rv == res.Value {
			return res
		}
		return b
	}
}

func ValueNodesReducer[T mcts.Counter, E comparable](r func(E, E) E) func(res, b ValueNodes[T, E]) ValueNodes[T, E] {
	return func(res, b ValueNodes[T, E]) ValueNodes[T, E] {
		rv := r(res.Value, b.Value)
		if rv == res.Value {
			return ValueNodes[T, E]{rv, append(res.Nodes, b.Nodes...)}
		}
		return b
	}
}
