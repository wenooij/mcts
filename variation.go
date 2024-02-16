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
type Variation []Node

// Root returns the StatEntry corresponding to the root.
//
// Root returns nil if the Variation is not rooted.
func (v Variation) Root() *Node {
	if len(v) == 0 || v[0].Action == nil {
		return nil
	}
	return &v[0]
}

// First returns the first StatEntry other than the root.
//
// First returns nil if the Variation is empty.
func (v Variation) First() *Node {
	if v = v.TrimRoot(); len(v) == 0 {
		return nil
	}
	return &v[0]
}

// Last returns the Last StatEntry for this variation.
//
// Last returns nil if the variation is empty.
func (v Variation) Last() *Node {
	if v = v.TrimRoot(); len(v) == 0 {
		return nil
	}
	leaf := v[len(v)-1]
	return &leaf
}

// TrimRoot returns the Variation v without its root node.
func (v Variation) TrimRoot() Variation {
	if len(v) == 0 || v[0].Action != nil {
		return v
	}
	return v[1:]
}

func (v Variation) String() (s string) {
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

// PV returns the principal variation for this Search.
//
// This is the line that the Search has searched deepest
// and is usually the best one.
//
// Use Stat to test arbitrary sequences.
func (r Search) PV() Variation { return r.FilterV(MaxRolloutsFilter(), AnyFilter(r.Rand)) }

// AnyV returns a random variation with runs for this Search.
//
// AnyV is useful for statistical sampling of the Search tree.
func (r Search) AnyV() Variation { return r.FilterV(AnyFilter(r.Rand)) }

// RootActions returns all actions searched from the root node.
//
// This may be a subset of the available RootActions.
func (r Search) RootActions() []Action {
	if r.Tree == nil {
		return nil
	}
	actions := make([]Action, 0, len(r.Tree.Children()))
	for _, child := range r.Tree.Children() {
		actions = append(actions, child.E.Action)
	}
	return actions
}

// Stat returns a sequence of Search stats for the given variation according to this Search.
//
// The returned Variation stops if the next action is not present in the Search tree.
func (r Search) Stat(vs ...Action) Variation {
	n := r.Tree
	if n == nil {
		return nil
	}
	res := make(Variation, 0, 1+len(vs))
	res = append(res, n.E)
	for _, s := range vs {
		child := getChild(n, s)
		if child == nil {
			// No existing child.
			break
		}
		// Add the StatEntry and continue down the line.
		n = child
		res = append(res, n.E)
	}
	return res
}

// InsertV merges a new variation into the search tree.
//
// Actions already present in the search have their scores added.
// Node priorities are recomputed using UCT.
//
// The Search is initialized if it had not already done so.
func (s *Search) InsertV(v Variation) {
	s.Init()
	n := s.Tree
	if root := v.Root(); root != nil {
		// Insert stat at Root.
	}
	for _, stat := range v.TrimRoot() {
		n, _ = getOrCreateChild(s, n, FrontierAction{
			Action: stat.Action,
			Weight: stat.PriorWeight,
		})
		e := n.E
		e.Score = stat.Score
		e.NumRollouts = stat.NumRollouts
	}
	// Fix priorities.
	backpropNull(n, s.ExploreFactor)
}
