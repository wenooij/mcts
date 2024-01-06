package mcts

import "github.com/wenooij/heapordered"

type topo[S Step] struct {
	parent *topo[S]
	*node[S]
	childSet map[S]*topo[S]
	children *heapordered.Tree[*topo[S]]
	Step     S
	depth    int
}

func (n *topo[S]) getNode() *node[S] {
	if n == nil {
		return nil
	}
	return n.node
}

func (n *topo[S]) getDepth() int {
	if n == nil {
		return 0
	}
	return n.depth
}

func (t *topo[S]) Init(parent *topo[S], n *node[S], step S) {
	*t = topo[S]{
		parent:   parent,
		node:     n,
		childSet: make(map[S]*topo[S]),
		children: heapordered.NewTree[*topo[S]](t),
		Step:     step,
		depth:    parent.getDepth(),
	}
	if parent != nil {
		t.parent.children.NewChildTree(t.children)
	}
}

func newTopoNode[S Step](s *Search[S], parent *topo[S], step S, log Log) *topo[S] {
	topo := new(topo[S])
	node := new(node[S])
	topo.Init(parent, node, step)
	node.Init(s, parent, topo, step, log)
	return topo
}

func (parent *topo[S]) newChild(s *Search[S], step S) (child *topo[S], created bool) {
	child, ok := parent.childSet[step]
	if ok {
		return child, false
	}
	child = newTopoNode(s, parent, step, s.Log())
	parent.childSet[step] = child
	return child, true
}
