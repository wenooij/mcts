package mcts

import "math/rand"

type topo[S Step] struct {
	parent *topo[S]
	*node[S]
	childSet map[S]*topo[S]
	children []*topo[S]
	Step     S
}

func (t *topo[S]) Init(parent *topo[S], n *node[S], step S) {
	t.parent = parent
	t.node = n
	t.childSet = make(map[S]*topo[S])
	t.children = make([]*topo[S], 0)
	t.Step = step
}

func newTopoNode[S Step](s *Search[S], si SearchInterface[S], parent *topo[S], step S, log Log, r *rand.Rand) *topo[S] {
	topo := new(topo[S])
	node := new(node[S])
	topo.Init(parent, node, step)
	node.Init(s, si, parent, topo, step, log, r)
	return topo
}

func (parent *topo[S]) newChild(s *Search[S], si SearchInterface[S], step S, r *rand.Rand) (child *topo[S], created bool) {
	child, ok := parent.childSet[step]
	if ok {
		return child, false
	}
	child = newTopoNode(s, si, parent, step, si.Log(), r)
	parent.childSet[step] = child
	parent.children = append(parent.children, child)
	return child, true
}
