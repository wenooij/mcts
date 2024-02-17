package mcts

import (
	"fmt"
	"math"
	"strings"

	"github.com/wenooij/heapordered"
)

type Node[T Counter] struct {
	Action      Action
	Score       Score[T]
	NumRollouts float64
	PriorWeight float64
}

// makeNode creates a tree node element.
func makeNode[T Counter](action FrontierAction) Node[T] {
	weight := action.Weight
	if weight < 0 {
		panic("mcts.Node: Predictor weight < 0 for step: " + action.Action.String())
	}
	if weight == 0 {
		weight = 1
	}
	return Node[T]{
		// Max priority for new nodes.
		// This will be recomputed after the first attempt.
		PriorWeight: weight,
		Action:      action.Action,
	}
}

func newTree[T Counter](s *Search[T]) *heapordered.Tree[Node[T]] {
	step := FrontierAction{}
	root := makeNode[T](step)
	return heapordered.NewTree(root, math.Inf(-1))
}

func createChild[T Counter](s *Search[T], parent *heapordered.Tree[Node[T]], action FrontierAction) (child *heapordered.Tree[Node[T]], created bool) {
	node := makeNode[T](action)
	child = heapordered.NewTree(node, math.Inf(-1))
	parent.NewChildTree(child)
	return child, true
}

func (e Node[T]) appendString(sb *strings.Builder) {
	fmt.Fprintf(sb, "[%f] %s (%d)", e.Score.Apply()/float64(e.NumRollouts), e.Action, int64(e.NumRollouts))
}

func (e Node[T]) String() string {
	var sb strings.Builder
	e.appendString(&sb)
	return sb.String()
}
