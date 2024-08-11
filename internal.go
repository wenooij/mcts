package mcts

import "math/rand"

// InternalInterface defines methods used for implementing the low level search procedure.
//
// Most users should use the builtin graph or tree API.
type InternalInterface[T Counter] struct {
	Backprop    func(trajectory []*Edge[T], counter CounterInterface[T], counters T, numRollouts, exploreFactor float64)
	Rollout     func(s SearchInterface[T], ri RolloutInterface[T], r *rand.Rand) (counters T, numRollouts float64)
	Expand      func(s SearchInterface[T], table map[uint64]*TableEntry[T], hashTable map[*TableEntry[T]]uint64, trajectory *[]*Edge[T], n *TableEntry[T], r *rand.Rand) (child *Edge[T])
	SelectChild func(s SearchInterface[T], table map[uint64]*TableEntry[T], hashTable map[*TableEntry[T]]uint64, trajectory *[]*Edge[T], n *TableEntry[T]) (child *Edge[T], expand bool)
	MakeNode    func(action FrontierAction) Node[T]
}
