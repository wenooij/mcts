package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type NodeType int

const (
	NodeTerminal NodeType = 1 << iota
)

type node[E Action] struct {
	NodeType      NodeType
	rawScore      Score
	numRollouts   float64
	exploreFactor float64
	weight        float64
	priority      float64

	// Topology.
	childSet map[E]*heapordered.Tree[*node[E]]
	Action   E
}

// newNode creates a new tree node element.
func newNode[E Action](step FrontierAction[E]) *node[E] {
	weight := step.Weight
	if weight < 0 {
		panic("mcts.Node: Predictor weight < 0 for step: " + step.Action.String())
	}
	if weight == 0 {
		weight = 1
	}
	return &node[E]{
		exploreFactor: step.ExploreFactor,
		// Max priority for new nodes.
		// This will be recomputed after the first attempt.
		priority: math.Inf(-1),
		weight:   weight,
		childSet: make(map[E]*heapordered.Tree[*node[E]]),
		Action:   step.Action,
	}
}

func (n *node[E]) Prioirty() float64 { return n.priority }

func (e *node[E]) RawScore() float64 {
	if e.rawScore == nil {
		return math.Inf(+1)
	}
	return e.rawScore.Score()
}

func (e *node[E]) NormScore() float64 {
	if e.numRollouts == 0 {
		return math.Inf(+1)
	}
	return e.rawScore.Score() / e.numRollouts
}

func newTree[E Action](s *Search[E]) *heapordered.Tree[*node[E]] {
	step := FrontierAction[E]{
		ExploreFactor: s.ExploreFactor,
	}
	root := newNode[E](step)
	return heapordered.NewTree(root)
}

func getOrCreateChild[E Action](s *Search[E], parent *heapordered.Tree[*node[E]], action FrontierAction[E]) (child *heapordered.Tree[*node[E]], created bool) {
	e := parent.Elem()
	if child, ok := e.childSet[action.Action]; ok {
		return child, false
	}
	if action.ExploreFactor == 0 {
		action.ExploreFactor = e.exploreFactor
	}
	node := newNode[E](action)
	child = parent.NewChild(node)
	e.childSet[action.Action] = child
	return child, true
}

func getChild[A Action](root *heapordered.Tree[*node[A]], action A) (child *heapordered.Tree[*node[A]]) {
	return root.Elem().childSet[action]
}
