package mcts

import "math/rand"

type node[S Step] struct {
	selector[S]
	expandBurnIn[S]
	expander[S]
	expandHeuristic
	rolloutRunner[S]
	backprop[S]
}

func (n *node[S]) Init(s *Search[S], si SearchInterface[S], parent, topo *topo[S], step S, log Log, r *rand.Rand) {
	n.selector.Init(topo, s.ExplorationParameter)
	n.expandBurnIn.Init(&n.expander, s.ExpandBurnInSamples)
	n.expander.Init(s, si, topo, &n.expandHeuristic, r)
	n.expandHeuristic.Init(s.MaxSpeculativeExpansions)
	n.backprop.Init(topo, si.Log())
}
