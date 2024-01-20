package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type node[S Step] struct {
	terminal      bool
	Log           Log
	numRollouts   float64
	exploreFactor float64
	priority      float64

	// Topology.
	childSet map[S]*heapordered.Tree[*node[S]]
	Step     S
}

func newNode[S Step](log Log, step FrontierStep[S]) *node[S] {
	return &node[S]{
		Log:           log,
		exploreFactor: step.ExploreFactor,
		priority:      step.Priority,
		childSet:      make(map[S]*heapordered.Tree[*node[S]]),
		Step:          step.Step,
	}
}

func (n *node[S]) Prioirty() float64 { return n.priority }

func (e *node[S]) Score() (float64, bool) {
	if e.Log == nil {
		return math.Inf(-1), false
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
	root := newNode[S](s.Log(), FrontierStep[S]{Step: sentinel, Priority: math.Inf(-1), ExploreFactor: s.ExploreFactor})
	return heapordered.NewTree(root)
}

func getOrCreateChild[S Step](s *Search[S], parent *heapordered.Tree[*node[S]], step FrontierStep[S]) (child *heapordered.Tree[*node[S]], created bool) {
	root, _ := parent.Elem()
	if child, ok := root.childSet[step.Step]; ok {
		return child, false
	}
	if step.ExploreFactor == 0 {
		step.ExploreFactor = root.exploreFactor
	}
	node := newNode[S](s.Log(), step)
	child = parent.NewChild(node)
	root.childSet[step.Step] = child
	return child, true
}

func getChild[S Step](root *heapordered.Tree[*node[S]], step S) (child *heapordered.Tree[*node[S]]) {
	e, _ := root.Elem()
	return e.childSet[step]
}
