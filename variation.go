package mcts

import (
	"fmt"
	"strings"
)

// Variation is a sequence of actions with Search statistics.
//
// The first element in the Variation may be a root node.
// It will have a nil Action as well among other differences.
// Use NodeType.Root to check or Variation.TrimRoot to trim it.
type Variation[T Counter] []Node[T]

// Root returns the StatEntry corresponding to the root.
//
// Root returns nil if the Variation is not rooted.
func (v Variation[T]) Root() *Node[T] {
	if len(v) == 0 || v[0].Action == nil {
		return nil
	}
	return &v[0]
}

// First returns the first StatEntry other than the root.
//
// First returns nil if the Variation is empty.
func (v Variation[T]) First() *Node[T] {
	if v = v.TrimRoot(); len(v) == 0 {
		return nil
	}
	return &v[0]
}

// Last returns the Last StatEntry for this variation.
//
// Last returns nil if the variation is empty.
func (v Variation[T]) Last() *Node[T] {
	if v = v.TrimRoot(); len(v) == 0 {
		return nil
	}
	leaf := v[len(v)-1]
	return &leaf
}

// TrimRoot returns the Variation v without its root node.
func (v Variation[T]) TrimRoot() Variation[T] {
	if len(v) == 0 || v[0].Action != nil {
		return v
	}
	return v[1:]
}

func (v Variation[T]) String() (s string) {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "[%f]", v[0].Score.Apply()/v[0].NumRollouts)
	for _, e := range v.TrimRoot() {
		fmt.Fprintf(&sb, " %s", e.Action.String())
	}
	fmt.Fprintf(&sb, " (%d)", int64(v[0].NumRollouts))
	return sb.String()
}
