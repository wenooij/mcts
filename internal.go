package mcts

import "math/rand"

// InternalInterface defines methods used for implementing the low level search procedure.
//
// Most users should use the builtin graph or tree API.
type InternalInterface[T Counter] struct {
	Init        func(s *Search[T])
	Reset       func(s *Search[T])
	Root        func()
	Backprop    func(counter CounterInterface[T], counters T, numRollouts, exploreFactor float64)
	Rollout     func(s SearchInterface[T], ri RolloutInterface[T], r *rand.Rand) (counters T, numRollouts float64)
	Expand      func(s SearchInterface[T], r *rand.Rand) (hasChild bool)
	SelectChild func(s SearchInterface[T]) (hasChild, expand bool)
	MakeNode    func(action FrontierAction) Node[T]
}
