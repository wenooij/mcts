package mcts

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultMaxSelectSamples     = 100
	defaultExplorationParameter = math.Sqrt2
)

// Search contains options used to run the MCTS Search.
//
// It also maintains a continuation which supports repeated calls to Search
// using the same search tree.
type Search[S Step] struct {
	root *topo[S]

	// SearchInterface implements the search space and steps for the search problem.
	SearchInterface[S]

	// Done signals the Search to stop when set.
	// When nil, the Search runs indefinitely.
	Done <-chan struct{}

	// Seed provides repeatable randomness to the search.
	// By default Seed is set to the current UNIX timestamp nanos.
	Seed int64

	// Rand provides randomness to the search.
	// If unset, it is automatically seeded based on the value from Seed.
	Rand *rand.Rand

	// ExpandBurnInSamples is the number of guaranteed Expand calls
	// applied initially before we start sampling from the node.
	// Default is 0.
	ExpandBurnInSamples int

	// MaxSpeculativeExpansions is the maximum number of speculative calls to Expand after optional burn in.
	// This applies a limit to the heuristic which calls Expand automatically in proportion to the hit-rate
	// of new steps. As the hit rate decreases, we call Expand less, up to this limit.
	// If set to 0, speculative samples are disabled, but be sure SelectBurnInSamples is nonzero to guarantee
	// Expand is called.
	// Default of 0 means no cap applied to speculative expansions.
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
	if s.ExplorationParameter == 0 {
		s.ExplorationParameter = defaultExplorationParameter
	}
	if s.Rand == nil {
		s.Rand = rand.New(rand.NewSource(s.Seed))
	}
}

// Reset deletes the search continuation so the next call to Search starts from scratch.
func (s *Search[S]) Reset() {
	s.root = nil
	s.Rand = nil
}

func (s *Search[S]) Search() Stat[S] {
	s.patchDefaults()
	if s.root == nil {
		var sentinel S
		s.root = newTopoNode(s, nil, sentinel, s.Log())
	}
	root := s.root
	for {
		node := root
		s.Root()
		for {
			next, ok := node.Select(s)
			if !ok {
				break
			}
			s.Apply(next.Step)
			node = next
		}
		frontier := node
		if expand := node.Expand(s); expand != nil {
			frontier = expand
			s.Apply(expand.Step)
		}
		frontier.Backprop(s.Rollout())
		select {
		case <-s.Done:
			return root.makeResult(s.Rand)
		default:
		}
	}
}
