package mcts

import (
	"math"
	"math/rand"
	"time"

	"github.com/wenooij/heapordered"
)

const defaultExploreFactor = math.Sqrt2

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

	// ExploreFactor is a tuneable parameter which weights the explore side of the
	// MAB policy.
	//
	// This should be made roughly proportional to values obtained from Log.Score.
	// Zero uses the default value of √2.
	ExploreFactor float64
}

func (s *Search[S]) patchDefaults() {
	if s.ExploreFactor == 0 {
		s.ExploreFactor = defaultExploreFactor
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

// Search runs the search until the Done channel is signalled.
//
// To run a deterministic number of epochs, call Init and SearchEpoch directly.
func (s *Search[S]) Search() {
	s.Init()
	for {
		// Run search epoches until done.
		s.SearchEpoch()
		select {
		case <-s.Done:
			return
		default:
		}
	}
}

// SearchEpoch runs a single epoch of MCTS.
func (s *Search[S]) SearchEpoch() {
	n := s.root
	s.Root() // Reset to root.
	// Select the best leaf node by MAB policy.
	for child := n.Min(); child != nil; n, child = child, child.Min() {
		s.Select(child.Elem().Step)
	}
	// Expand the leaf node.
	if child := expand(s, n); child != nil {
		s.Select(child.Elem().Step)
		n = child
	}
	log, numRollouts := s.Rollout()        // Simulate rollouts.
	backprop(n, log, float64(numRollouts)) // Backprop the Score.
}
