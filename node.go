package mcts

type node[S Step] struct {
	terminal     bool
	Log          Log
	numRollouts  float64
	exploreParam float64
	priority     float64

	// Speculative expansion.
	hits, samples int
	burnedIn      bool
}

func (n *node[S]) Init(s *Search[S], parent, topo *topo[S], step S, log Log) {
	n.Log = s.Log()
	n.exploreParam = s.ExplorationParameter
	n.priority = s.InitialNodePriority
}

func (n *node[S]) Prioirty() float64 { return n.priority }
