package mcts

import (
	"fmt"
	"strings"
)

// TableEntry represents an adjacency list of Edges available from this node.
type TableEntry[T Counter] []*Edge[T]

type Edge[T Counter] struct {
	Src, Dst *TableEntry[T]
	Priority float64
	Node[T]
}

type Node[T Counter] struct {
	Action      Action
	Score       Score[T]
	NumRollouts float64
	PriorWeight float64
}

func (e Node[T]) appendString(sb *strings.Builder) {
	fmt.Fprintf(sb, "[%f] %s (%d)", e.Score.Apply()/float64(e.NumRollouts), e.Action, int64(e.NumRollouts))
}

func (e Node[T]) String() string {
	var sb strings.Builder
	e.appendString(&sb)
	return sb.String()
}
