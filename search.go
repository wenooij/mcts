package mcts

import (
	"math/rand"
	"time"

	"github.com/wenooij/heapordered"
)

// DefaultExploreFactor based on the theory assuming scores normalized to the interval [-1, +1].
//
// In practice, ExploreFactor is a tunable hyperparameter.
const DefaultExploreFactor = 1.224744871391589 // √3/√2

// Search contains options used to run the MCTS Search.
//
// It also maintains a continuation which supports repeated calls to Search
// using the same search tree.
//
// Many of the hyperparameters have drastic impacts on Search performance and need
// to be experimentally tuned first. See FitParams in the model subpackage for more info.
type Search struct {
	*heapordered.Tree[Node]

	// SearchInterface implements the search environment.
	SearchInterface

	// NumEpisodes ends the Search after the given fixed number
	// of episodes. Default is 100.
	NumEpisodes int

	// Seed provides repeatable randomness to the search.
	// By default Seed is set to the current UNIX timestamp nanos.
	Seed int64

	// Rand provides randomness to the search.
	// If unset, it is automatically seeded based on the value from Seed.
	Rand *rand.Rand

	// ExploreFactor is a tuneable parameter which weights the explore side of the
	// MAB policy.
	//
	// ExploreFactor may be changed during between calls to Search effectively
	// allowing it to behave like a temperature parameter as in simulated annealing.
	// It may be helpful to start at a higher ExploreFactor and gradually decrease it
	// as the search matures.
	//
	// This should be made roughly proportional to scores obtained from random rollouts.
	// Zero uses the default value of DefaultExploreFactor.
	ExploreFactor float64
}

func (s *Search) patchDefaults() {
	if s.ExploreFactor == 0 {
		s.ExploreFactor = DefaultExploreFactor
	}
	if s.NumEpisodes == 0 {
		s.NumEpisodes = 100
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
func (s *Search) Init() bool {
	if s.Tree != nil {
		return false
	}
	s.patchDefaults()
	s.Tree = newTree(s)
	initializeScore(s, s.Tree)
	return true
}

// Reset deletes the search continuation and RNG so the next call to Search starts from scratch.
func (s *Search) Reset() {
	s.Tree = nil
	s.Rand = nil
}

// Search runs the search until the Done channel is signalled.
//
// To run a deterministic number of runs, set FixedEpisodes.
func (s *Search) Search() {
	s.Init()
	for i := 0; i < s.NumEpisodes; i++ {
		s.searchEpisode()
	}
}

func (s *Search) searchEpisode() {
	n := s.Tree
	s.SearchInterface.Root() // Reset to root.
	// Select the best leaf node by MAB policy.
	for child := selectChild(s, n); child != nil; n, child = child, selectChild(s, child) {
	}
	// Expand a new frontier node.
	if frontier := expand(s, n); frontier != nil {
		n = frontier
	}
	// Simulate and backprop score.
	if rawScore, numRollouts := rollout(s, n); numRollouts != 0 {
		backprop(n, rawScore, numRollouts, s.ExploreFactor)
	}
}
