package mcts

import (
	"math"

	"github.com/wenooij/heapordered"
)

type nodeType int

const (
	nodeTerminal nodeType = 1 << iota
	nodeRoot
	nodePartial
	nodeExpanded
)

func (n nodeType) Terminal() bool       { return n&nodeTerminal != 0 }
func (n nodeType) Root() bool           { return n&nodeRoot != 0 }
func (n nodeType) PartlyExpanded() bool { return n&nodePartial != 0 }
func (n nodeType) Expanded() bool       { return n&nodeExpanded != 0 }

type Node struct {
	nodeType
	rawScore          Score
	numParentRollouts float64
	numRollouts       float64
	exploreFactor     float64
	predictWeight     float64
	priority          float64

	// Topology.
	depth  int
	action Action
}

// newNode creates a new tree node element.
func newNode(action FrontierAction) *Node {
	weight := action.Weight
	if weight < 0 {
		panic("mcts.Node: Predictor weight < 0 for step: " + action.Action.String())
	}
	if weight == 0 {
		weight = 1
	}
	return &Node{
		exploreFactor: action.ExploreFactor,
		// Max priority for new nodes.
		// This will be recomputed after the first attempt.
		priority:      math.Inf(-1),
		predictWeight: weight,
		action:        action.Action,
	}
}

func (n *Node) Action() Action             { return n.action }
func (n *Node) NumRollouts() float64       { return n.numRollouts }
func (n *Node) PredictWeight() float64     { return n.predictWeight }
func (n *Node) ExploreFactor() float64     { return n.exploreFactor }
func (n *Node) NumParentRollouts() float64 { return n.numParentRollouts }
func (n *Node) Priority() float64          { return float64(n.priority) }
func (n *Node) RawScore() Score            { return n.rawScore }
func (n *Node) Depth() int                 { return n.depth }
func (n *Node) Score() float64 {
	if n.numRollouts == 0 {
		return math.Inf(+1)
	}
	return n.rawScore.Score() / n.numRollouts
}
func (n *Node) rawScoreValue() float64 {
	if n.rawScore == nil {
		return math.Inf(+1)
	}
	return n.rawScore.Score()
}

func newTree(s *Search) *heapordered.Tree[*Node] {
	step := FrontierAction{
		ExploreFactor: s.ExploreFactor,
	}
	root := newNode(step)
	root.nodeType |= nodeRoot
	return heapordered.NewTree(root)
}

func getOrCreateChild(s *Search, parent *heapordered.Tree[*Node], action FrontierAction) (child *heapordered.Tree[*Node], created bool) {
	e := parent.Elem()
	if e.PartlyExpanded() {
		// For partly expanded nodes we need to use the slow check
		// to avoid wasting resources creating duplicate children.
		if child := getChild(parent, action.Action); child != nil {
			return child, false
		}
	}
	if action.ExploreFactor == 0 {
		action.ExploreFactor = e.exploreFactor
	}
	node := newNode(action)
	node.depth = e.depth + 1
	child = parent.NewChild(node)
	return child, true
}

func getChild(root *heapordered.Tree[*Node], action Action) (child *heapordered.Tree[*Node]) {
	for _, e := range root.Children() {
		if e.Elem().action == action {
			return e
		}
	}
	return nil
}
