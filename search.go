package mcts

import (
	"fmt"
	"math"
	"math/rand"
	"plugin"
	"time"
)

// DefaultExploreFactor based on the theory assuming scores normalized to the interval [-1, +1].
//
// In practice, ExploreFactor is a tunable hyperparameter.
const DefaultExploreFactor = math.Sqrt2

// Search contains options used to run the MCTS Search.
//
// It also maintains a continuation which supports repeated calls to Search
// using the same search tree.
//
// Many of the hyperparameters have drastic impacts on Search performance and need
// to be experimentally tuned first. See FitParams in the model subpackage for more info.
type Search[T Counter] struct {
	// SearchInterface implements the search environment.
	SearchInterface[T]

	// Optional counter implementation.
	CounterInterface[T]

	RootEntry *EdgeList[T]

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

func (s *Search[T]) patchDefaults() {
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
	if s.CounterInterface.Add == nil {
		patchBuiltinAdd[T](&s.CounterInterface)
	}
}

func (s *Search[T]) LoadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}
	sym, err := p.Lookup("SearchInterface")
	if err != nil {
		return err
	}
	impl, ok := sym.(SearchInterface[T])
	if !ok {
		return fmt.Errorf("unexpected exported type for SearchInterface")
	}
	s.SearchInterface = impl
	return nil
}

// Init create a new root for the search if it doesn't exist yet.
// Init additionally patches default parameter values.
func (s *Search[T]) Init() bool {
	s.patchDefaults()
	s.InternalInterface.Init(s)
	if s.SearchInterface.Root == nil {
		panic("Search.Init: Search.SearchInterface.Root is nil. A search implementation is required before calling Search or Init.")
	}
	return true
}

// Reset deletes the search continuation and RNG so the next call to Search starts from scratch.
func (s *Search[T]) Reset() {
	s.InternalInterface.Reset(s)
	s.Rand = nil
}

// Search runs the search NumEpisodes times.
func (s *Search[T]) Search() {
	s.Init()
	for i := 0; i < s.NumEpisodes; i++ {
		s.searchEpisode()
	}
}

func (s *Search[T]) searchEpisode() {
	n := s.RootEntry
	s.InternalInterface.Root()
	s.SearchInterface.Root() // Reset to root.
	// Select the best leaf node by MAB policy.
	var doExpand bool
	for child := (*Edge[T])(nil); ; n = child.Dst {
		if child, doExpand = s.SelectChild(s.SearchInterface, n); child == nil {
			break
		}
	}
	// Expand a new frontier node.
	if doExpand {
		s.InternalInterface.Expand(s.SearchInterface, n, s.Rand)
	}
	// Simulate and backprop score.
	if counters, numRollouts := s.InternalInterface.Rollout(s.SearchInterface, s.RolloutInterface, s.Rand); numRollouts != 0 {
		s.Backprop(s.CounterInterface, counters, numRollouts, s.ExploreFactor)
	}
}
