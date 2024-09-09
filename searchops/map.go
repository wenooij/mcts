package searchops

import (
	"github.com/wenooij/mcts"
)

// MinMax records min and max floating point values.
type MinMax struct {
	Min float64
	Max float64
}

func MinMaxMapper[T mcts.Counter](m func(mcts.Node[T]) float64) func(mcts.Node[T]) MinMax {
	return func(n mcts.Node[T]) MinMax {
		v := m(n)
		return MinMax{v, v}
	}
}

type ValueNode[T mcts.Counter, E comparable] struct {
	Value E
	Node  mcts.Node[T]
}

func ValueNodeMapper[T mcts.Counter, E comparable](m func(mcts.Node[T]) E) func(mcts.Node[T]) ValueNode[T, E] {
	return func(n mcts.Node[T]) ValueNode[T, E] {
		return ValueNode[T, E]{m(n), n}
	}
}

type ValueNodes[T mcts.Counter, E comparable] struct {
	Value E
	Nodes []mcts.Node[T]
}

func ValueNodesMapper[T mcts.Counter, E comparable](m func(mcts.Node[T]) E) func(mcts.Node[T]) ValueNodes[T, E] {
	return func(n mcts.Node[T]) ValueNodes[T, E] {
		return ValueNodes[T, E]{m(n), []mcts.Node[T]{n}}
	}
}

func ActionMapper[T mcts.Counter]() func(mcts.Node[T]) mcts.Action {
	return func(n mcts.Node[T]) mcts.Action { return n.Action }
}

func RawScoreMapper[T mcts.Counter]() func(mcts.Node[T]) mcts.Score[T] {
	return func(n mcts.Node[T]) mcts.Score[T] { return n.Score }
}

func ScoreMapper[T mcts.Counter]() func(mcts.Node[T]) float64 {
	return func(n mcts.Node[T]) float64 { return n.Score.Objective(n.Score.Counter) }
}

func PriorityMapper[T mcts.Counter]() func(mcts.Node[T]) float64 {
	return func(n mcts.Node[T]) float64 { return n.Priority }
}

func RolloutsMapper[T mcts.Counter]() func(mcts.Node[T]) float64 {
	return func(n mcts.Node[T]) float64 { return n.NumRollouts }
}

func WeightMapper[T mcts.Counter]() func(mcts.Node[T]) float64 {
	return func(n mcts.Node[T]) float64 { return n.PriorWeight }
}
