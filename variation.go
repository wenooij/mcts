package mcts

import (
	"fmt"
	"strings"
)

// Variation is a sequence of actions with Search statistics.
//
// The first element in the Variation is the root mode and will have the zero value for the Step.
type Variation []StatEntry

func (v Variation) Leaf() *StatEntry {
	if len(v) == 0 {
		return nil
	}
	leaf := v[len(v)-1]
	return &leaf
}

func (v Variation) String() string {
	var sb strings.Builder
	if len(v) == 0 {
		return "\n"
	}
	for i := 0; i < len(v); i++ {
		e := v[i]
		fmt.Fprintf(&sb, "%-2d ", i)
		e.appendString(&sb)
		fmt.Fprintln(&sb)
	}
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

// Stat returns a sequence of Search stats for the given variation according to this Search.
func (r Search) Stat(vs ...Action) Variation {
	n := r.root
	if n == nil {
		return nil
	}
	res := make(Variation, 0, 1+len(vs))
	res = append(res, makeStatEntry(n))
	for _, s := range vs {
		child := getChild(n, s)
		if child == nil {
			// No existing child.
			break
		}
		// Add the StatEntry and continue down the line.
		n = child
		res = append(res, makeStatEntry(n))
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
	n := s.root
	for _, stat := range v {
		var created bool
		n, created = getOrCreateChild(s, n, FrontierAction{
			Action:        stat.Action,
			Weight:        stat.PredictorWeight,
			ExploreFactor: stat.ExploreFactor,
		})
		e := n.Elem()
		if created {
			e.nodeType = stat.NodeType
		} else {
			e.nodeType &= ^nodeTerminal          // Clear the terminal bit.
			e.exploreFactor = stat.ExploreFactor // Reset the explore factor.
		}
		e.rawScore = stat.RawScore
		e.numRollouts += stat.NumRollouts
	}
	// Fix priorities.
	backpropNull(n)
}
