package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type node[S Step] struct {
	terminal     bool
	Log          Log
	numRollouts  float64
	exploreParam float64
	priority     float64

	// Speculative expansion.
	hits, samples int
	burnedIn      bool

	// Topology.
	childSet map[S]*heapordered.Tree[*node[S]]
	Step     S
}

func newNode[S Step](s *Search[S], step S) *node[S] {
	return &node[S]{
		Log:          s.Log(),
		exploreParam: s.ExplorationParameter,
		priority:     s.NodePolicy(step),
		childSet:     make(map[S]*heapordered.Tree[*node[S]]),
		Step:         step,
	}
}

func (n *node[S]) Prioirty() float64 { return n.priority }

func (e *node[S]) Score() (float64, bool) {
	if e.Log == nil {
		return 0, false
	}
	return e.Log.Score(), true
}

func (e *node[S]) NormScore() float64 {
	if score, ok := e.Score(); ok && e.numRollouts > 0 {
		return score / e.numRollouts
	} else {
		return math.Inf(-1)
	}
}

func newTree[S Step](s *Search[S]) *heapordered.Tree[*node[S]] {
	var sentinel S
	root := newNode[S](s, sentinel)
	return heapordered.NewTree(root)
}

func getOrCreateChild[S Step](s *Search[S], parent *heapordered.Tree[*node[S]], step S) (child *heapordered.Tree[*node[S]], created bool) {
	root, _ := parent.Elem()
	if child, ok := root.childSet[step]; ok {
		return child, false
	}
	node := newNode[S](s, step)
	child = parent.NewChild(node)
	root.childSet[step] = child
	return child, true
}

func getChild[S Step](root *heapordered.Tree[*node[S]], step S) (child *heapordered.Tree[*node[S]]) {
	e, _ := root.Elem()
	return e.childSet[step]
}
