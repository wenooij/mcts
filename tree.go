package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type NodeType int

const (
	nodeTerminal NodeType = 1 << iota
)

func (t NodeType) Terminal() bool { return t&nodeTerminal != 0 }

type node struct {
	nodeType      NodeType
	rawScore      Score
	numRollouts   float64
	exploreFactor float64
	weight        float64
	priority      float64

	// Topology.
	childSet map[Action]*heapordered.Tree[*node]
	Action   Action
}

// newNode creates a new tree node element.
func newNode(action FrontierAction) *node {
	weight := action.Weight
	if weight < 0 {
		panic("mcts.Node: Predictor weight < 0 for step: " + action.Action.String())
	}
	if weight == 0 {
		weight = 1
	}
	return &node{
		exploreFactor: action.ExploreFactor,
		// Max priority for new nodes.
		// This will be recomputed after the first attempt.
		priority: math.Inf(-1),
		weight:   weight,
		childSet: make(map[Action]*heapordered.Tree[*node]),
		Action:   action.Action,
	}
}

func (n *node) Prioirty() float64 { return n.priority }

func (e *node) RawScore() float64 {
	if e.rawScore == nil {
		return math.Inf(+1)
	}
	return e.rawScore.Score()
}

func (e *node) NormScore() float64 {
	if e.numRollouts == 0 {
		return math.Inf(+1)
	}
	return e.rawScore.Score() / e.numRollouts
}

func newTree(s *Search) *heapordered.Tree[*node] {
	step := FrontierAction{
		ExploreFactor: s.ExploreFactor,
	}
	root := newNode(step)
	return heapordered.NewTree(root)
}

func getOrCreateChild(s *Search, parent *heapordered.Tree[*node], action FrontierAction) (child *heapordered.Tree[*node], created bool) {
	e := parent.Elem()
	if child, ok := e.childSet[action.Action]; ok {
		return child, false
	}
	if action.ExploreFactor == 0 {
		action.ExploreFactor = e.exploreFactor
	}
	node := newNode(action)
	child = parent.NewChild(node)
	e.childSet[action.Action] = child
	return child, true
}

func getChild(root *heapordered.Tree[*node], action Action) (child *heapordered.Tree[*node]) {
	return root.Elem().childSet[action]
}
