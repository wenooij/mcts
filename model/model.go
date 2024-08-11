package model

import (
	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/internal/graph"
)

func MakeSearchInterface[T mcts.Counter](x any, counter mcts.CounterInterface[T]) mcts.SearchInterface[T] {
	var hash func() uint64
	if h, ok := x.(interface{ Hash() uint64 }); ok {
		hash = h.Hash
	}
	return mcts.SearchInterface[T]{
		Root:   x.(interface{ Root() }).Root,
		Select: x.(interface{ Select(mcts.Action) bool }).Select,
		Expand: x.(interface {
			Expand(int) []mcts.FrontierAction
		}).Expand,
		Score:             x.(interface{ Score() mcts.Score[T] }).Score,
		Hash:              hash,
		RolloutInterface:  makeRolloutInterface[T](x),
		CounterInterface:  counter,
		InternalInterface: graph.InternalInterface[T](),
	}
}

func makeRolloutInterface[T mcts.Counter](x any) mcts.RolloutInterface[T] {
	ri, ok := x.(interface{ Rollout() (T, float64) })
	if !ok {
		return mcts.RolloutInterface[T]{}
	}
	return mcts.RolloutInterface[T]{Rollout: ri.Rollout}
}
