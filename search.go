package mcts

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"

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
type Search[T Counter] struct {
	*heapordered.Tree[Node[T]]

	// SearchInterface implements the search environment.
	SearchInterface[T]

	// RolloutInterface provides the optional custom rollout implementation.
	RolloutInterface[T]

	// NumEpisodes ends the Search after the given fixed number
	// of episodes. Default is 100.
	NumEpisodes int

	// Seed provides repeatable randomness to the search.
	// By default Seed is set to the current UNIX timestamp nanos.
	Seed int64

	// Rand provides randomness to the search.
	// If unset, it is automatically seeded based on the value from Seed.
	Rand *rand.Rand

	// AddCounters is provided to supply a generic function to add the results of
	// the backpropagated scores from search.
	//
	// For common values of T this will be populated automatically otherwise you
	// will need to supply a function yourself. The common values of T are:
	// float32, float64, [2]float64, []float64, int, and int64.
	// AddCounters for counters in the model package need to be supplied manually.
	AddCounters func(T, T) T

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
	if s.RolloutInterface == nil {
		if rolloutInterface, ok := s.SearchInterface.(RolloutInterface[T]); ok {
			s.RolloutInterface = rolloutInterface
		}
	}
	if s.AddCounters == nil {
		var t T
		switch any(t).(type) {
		case float32:
			f := func(x1, x2 float32) float32 { return x1 + x2 }
			s.AddCounters = *(*func(T, T) T)(unsafe.Pointer(&f))
		case float64:
			f := func(x1, x2 float64) float64 { return x1 + x2 }
			s.AddCounters = *(*func(T, T) T)(unsafe.Pointer(&f))
		case [2]float64:
			f := func(c1, c2 [2]float64) [2]float64 { c1[0] += c2[0]; c1[1] += c2[1]; return c1 }
			s.AddCounters = *(*func(T, T) T)(unsafe.Pointer(&f))
		case []float64:
			f := func(c1, c2 []float64) []float64 {
				for i, v := range c2 {
					c1[i] += v
				}
				return c1
			}
			s.AddCounters = *(*func(T, T) T)(unsafe.Pointer(&f))
		case int:
			f := func(x1, x2 int) int { return x1 + x2 }
			s.AddCounters = *(*func(T, T) T)(unsafe.Pointer(&f))
		case int64:
			f := func(x1, x2 int64) int64 { return x1 + x2 }
			s.AddCounters = *(*func(T, T) T)(unsafe.Pointer(&f))
		default:
			panic(fmt.Errorf("Search.Init: could not automatically determine a value for AddCounters for type %T: must supply a custom function", t))
		}
	}
}

// Init create a new root for the search if it doesn't exist yet.
// Init additionally patches default parameter values.
func (s *Search[T]) Init() bool {
	if s.Tree != nil {
		return false
	}
	s.patchDefaults()
	s.Tree = newTree(s)
	initializeScore(s, s.Tree)
	return true
}

// Reset deletes the search continuation and RNG so the next call to Search starts from scratch.
func (s *Search[T]) Reset() {
	s.Tree = nil
	s.Rand = nil
}

// Search runs the search until the Done channel is signalled.
//
// To run a deterministic number of runs, set FixedEpisodes.
func (s *Search[T]) Search() {
	s.Init()
	for i := 0; i < s.NumEpisodes; i++ {
		s.searchEpisode()
	}
}

func (s *Search[T]) searchEpisode() {
	n := s.Tree
	s.SearchInterface.Root() // Reset to root.
	// Select the best leaf node by MAB policy.
	var doExpand bool
	for child := (*heapordered.Tree[Node[T]])(nil); ; n = child {
		if child, doExpand = selectChild(s, n); child == nil {
			break
		}
	}
	// Expand a new frontier node.
	if doExpand {
		if frontier := expand(s, n); frontier != nil {
			n = frontier
		}
	}
	// Simulate and backprop score.
	if counters, numRollouts := rollout(s, n); numRollouts != 0 {
		backprop(n, s.AddCounters, counters, numRollouts, s.ExploreFactor)
	}
}
