package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type Node struct {
	action            Action
	rawScore          Score
	numParentRollouts float64
	numRollouts       float64
	predictWeight     float64
}

// makeNode creates a tree node element.
func makeNode(action FrontierAction) Node {
	weight := action.Weight
	if weight < 0 {
		panic("mcts.Node: Predictor weight < 0 for step: " + action.Action.String())
	}
	if weight == 0 {
		weight = 1
	}
	return Node{
		// Max priority for new nodes.
		// This will be recomputed after the first attempt.
		predictWeight: weight,
		action:        action.Action,
	}
}

func (n Node) Action() Action             { return n.action }
func (n Node) NumRollouts() float64       { return n.numRollouts }
func (n Node) PredictWeight() float64     { return n.predictWeight }
func (n Node) NumParentRollouts() float64 { return n.numParentRollouts }
func (n Node) RawScore() Score            { return n.rawScore }
func (n Node) Score() float64 {
	if n.numRollouts == 0 {
		return math.Inf(+1)
	}
	return n.rawScore.Apply() / n.numRollouts
}
func (n Node) rawScoreValue() float64 { return n.rawScore.Apply() }

func newTree(s *Search) *heapordered.Tree[Node] {
	step := FrontierAction{}
	root := makeNode(step)
	return heapordered.NewTree(root, math.Inf(-1))
}

func getOrCreateChild(s *Search, parent *heapordered.Tree[Node], action FrontierAction) (child *heapordered.Tree[Node], created bool) {
	node := makeNode(action)
	child = heapordered.NewTree(node, math.Inf(-1))
	parent.NewChildTree(child)
	return child, true
}

func getChild(root *heapordered.Tree[Node], action Action) (child *heapordered.Tree[Node]) {
	for _, e := range root.Children() {
		if e.E.action == action {
			return e
		}
	}
	return nil
}
