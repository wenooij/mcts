package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type NodeType int

const (
	NodeTerminal NodeType = 1 << iota
)

type node[S Step] struct {
	NodeType      NodeType
	rawScore      Score
	numRollouts   float64
	exploreFactor float64
	weight        float64
	priority      float64

	// Topology.
	childSet map[S]*heapordered.Tree[*node[S]]
	Step     S
}

// newNode creates a new tree node element.
func newNode[S Step](step FrontierStep[S]) *node[S] {
	weight := step.Weight
	if weight < 0 {
		panic("mcts.Node: Predictor weight < 0 for step: " + step.Step.String())
	}
	if weight == 0 {
		weight = 1
	}
	return &node[S]{
		exploreFactor: step.ExploreFactor,
		// Max priority for new nodes.
		// This will be recomputed after the first attempt.
		priority: math.Inf(-1),
		weight:   weight,
		childSet: make(map[S]*heapordered.Tree[*node[S]]),
		Step:     step.Step,
	}
}

func (n *node[S]) Prioirty() float64 { return n.priority }

func (e *node[S]) RawScore() float64 {
	if e.rawScore == nil {
		return math.Inf(+1)
	}
	return e.rawScore.Score()
}

func (e *node[S]) NormScore() float64 {
	if e.numRollouts == 0 {
		return math.Inf(+1)
	}
	return e.rawScore.Score() / e.numRollouts
}

func newTree[S Step](s *Search[S]) *heapordered.Tree[*node[S]] {
	var sentinel S
	step := FrontierStep[S]{
		Step:          sentinel,
		ExploreFactor: s.ExploreFactor,
	}
	root := newNode[S](step)
	return heapordered.NewTree(root)
}

func getOrCreateChild[S Step](s *Search[S], parent *heapordered.Tree[*node[S]], step FrontierStep[S]) (child *heapordered.Tree[*node[S]], created bool) {
	e := parent.Elem()
	if child, ok := e.childSet[step.Step]; ok {
		return child, false
	}
	if step.ExploreFactor == 0 {
		step.ExploreFactor = e.exploreFactor
	}
	node := newNode[S](step)
	child = parent.NewChild(node)
	e.childSet[step.Step] = child
	return child, true
}

func getChild[S Step](root *heapordered.Tree[*node[S]], step S) (child *heapordered.Tree[*node[S]]) {
	return root.Elem().childSet[step]
}
