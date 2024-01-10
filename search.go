package mcts

import (
	"math"
	"math/rand"
	"time"

	"github.com/wenooij/heapordered"
)

const (
	defaultExpandBufferSize     = 64
	defaultMaxSelectSamples     = 100
	defaultExplorationParameter = math.Sqrt2
)

// Search contains options used to run the MCTS Search.
//
// It also maintains a continuation which supports repeated calls to Search
// using the same search tree.
//
// Many of the hyperparameters have drastic impacts on Search performance and need
// to be experimentally tuned first. See FitParams in the model subpackage for more info.
type Search[S Step] struct {
	root *heapordered.Tree[*node[S]]

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

	// ExpandBufferSize is the size of the steps buffer passed to Expand in the SearchInferface.
	// Default is 64.
	ExpandBufferSize int

	expandBuffer []FrontierStep[S]

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
	if s.ExpandBufferSize == 0 {
		s.ExpandBufferSize = defaultExpandBufferSize
	}
	if cap(s.expandBuffer) != s.ExpandBufferSize {
		s.expandBuffer = make([]FrontierStep[S], s.ExpandBufferSize)
	}
	if s.ExplorationParameter == 0 {
		s.ExplorationParameter = defaultExplorationParameter
	}
	if s.Rand == nil {
		if s.Seed == 0 {
			s.Seed = time.Now().UnixNano()
		}
		s.Rand = rand.New(rand.NewSource(s.Seed))
	}
}

// Init create a new root for the search if it doesn't exist yet.
// Init additionally patches default parameter values.
func (s *Search[S]) Init() bool {
	if s.root != nil {
		return false
	}
	s.patchDefaults()
	s.root = newTree(s)
	return true
}

// Reset deletes the search continuation and RNG so the next call to Search starts from scratch.
func (s *Search[S]) Reset() {
	s.root = nil
	s.Rand = nil
}

// Search runs the search until Done is signalled.
//
// If reproducible results are required, use Init and SearchEpoch directly.
func (s *Search[S]) Search() {
	s.Init()
	for {
		s.SearchEpoch()
		select {
		case <-s.Done:
			// Done signal. Complete the Search.
			return
		default:
		}
	}
}

// SearchEpoch runs a single epoch of search.
func (s *Search[S]) SearchEpoch() {
	n := s.root
	s.Root()
	for {
		child, ok := selectChild(s, n)
		if !ok {
			break
		}
		e, _ := child.Elem()
		s.Select(e.Step)
		n = child
	}
	if expand := expand(s, n); expand != nil {
		n = expand
		e, _ := n.Elem()
		s.Select(e.Step)
	}
	frontier := n
	log, numRollouts := s.Rollout()
	backprop(frontier, log, numRollouts)
}
