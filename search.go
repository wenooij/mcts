package mcts

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultRolloutsPerEpoch         = 3
	defaultMaxSelectSamples         = 100
	defaultMaxSpeculativeExpansions = 100
	defaultExplorationParameter     = math.Sqrt2
)

// Search contains options used to run the MCTS Search.
//
// It also maintains a continuation which supports repeated calls to Search
// using the same search tree.
type Search[S Step] struct {
	root *topo[S]

	// Seed provides repeatable randomness to the search.
	// By default Seed is set to the current UNIX timestamp nanos.
	Seed int64

	// ExpandBurnInSamples is the number of guaranteed Expand calls
	// applied initially before we start sampling from the node.
	// Default is 0.
	ExpandBurnInSamples int

	// RolloutsPerEpoch is a tuneable number of calls to Rollout per Search epoch.
	// This value should be high enough to amortize the cost of selecting a node
	// but not too high that it would take too much time away from other searches.
	// Default is 100.
	RolloutsPerEpoch int

	// MinExpandDepth is the minimum depth in which a rollout is allowed.
	// This is useful for search which need a few steps to get set up,
	// or when rolling out before MinExpandDepth is not well defined.
	// MinExpandDepth doesn't apply if Expand returns an empty step.
	// Default is 0.
	MinExpandDepth int

	// MaxSpeculativeExpansions is the maximum number of speculative calls to Expand after optional burn in.
	// This applies a limit to the heuristic which calls Expand automatically in proportion to the hit-rate
	// of new steps. As the hit rate decreases, we call Expand less, up to this limit.
	// If set to 0, speculative samples are disabled, but be sure SelectBurnInSamples is nonzero to guarantee
	// Expand is called.
	// Default is 100.
	MaxSpeculativeExpansions int

	// ExplorationParameter is a tuneable parameter which weights the explore side of the
	// MAB policy.
	// Zero will use the default value of √2.
	ExplorationParameter float64
}

func (s *Search[S]) patchDefaults() {
	if s.Seed == 0 {
		s.Seed = time.Now().UnixNano()
	}
	if s.RolloutsPerEpoch == 0 {
		s.RolloutsPerEpoch = defaultRolloutsPerEpoch
	}
	if s.MaxSpeculativeExpansions == 0 {
		s.MaxSpeculativeExpansions = defaultMaxSpeculativeExpansions
	}
	if s.ExplorationParameter == 0 {
		s.ExplorationParameter = defaultExplorationParameter
	}
}

// Reset deletes the search continuation so the next call to Search starts from scratch.
func (s *Search[S]) Reset() {
	s.root = nil
}

func (s *Search[S]) Search(si SearchInterface[S], done <-chan struct{}) Stat[S] {
	s.patchDefaults()
	r := rand.New(rand.NewSource(s.Seed))
	if s.root == nil {
		var sentinel S
		s.root = newTopoNode(s, si, nil, sentinel, si.Log(), r)
	}
	root := s.root
	for {
		node := root
		si.Root()
		for {
			next, ok := node.Select()
			if !ok {
				break
			}
			si.Apply(next.Step)
			node = next
		}
		frontier := node
		if expand := node.Expand(); expand != nil {
			frontier = expand
			si.Apply(expand.Step)
		}
		frontier.Backprop(frontier.Rollout())
		select {
		case <-done:
			return root.makeResult(r)
		default:
		}
	}
}
