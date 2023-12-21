package mcts

import (
	"math"
	"math/rand"
	"time"
)

const (
	defaultRolloutsPerEpoch     = 3
	defaultMaxSelectSamples     = 100
	defaultExplorationParameter = math.Sqrt2
)

// Search contains options used to run the MCTS Search.
//
// It also maintains a continuation which supports repeated calls to Search
// using the same search tree.
type Search[E Step] struct {
	root *EventLog[E]

	// Seed provides repeatable randomness to the search.
	// By default Seed is set to the current UNIX timestamp nanos.
	Seed int64

	// MinSelectDepth is the minimum depth in which a rollout is allowed.
	// This is useful for search which need a few steps to get set up,
	// or when rolling out before MinSelectDepth is not well defined.
	// MinSelectDepth doesn't apply if Expand returns an empty step.
	// Default is 0.
	MinSelectDepth int

	// SelectBurnInSamples is the number of guaranteed Expand calls
	// applied initially before we start sampling from the node.
	// Default is 0.
	SelectBurnInSamples int

	// MaxSpeculativeSamples is the maximum number of speculative calls to Expand after optional burn in.
	// This applies a limit to the heuristic which calls Expand automatically in proportion to the hit-rate
	// of new steps. As the hit rate decreases, we call Expand less, up to this limit.
	// If set to 0, speculative samples are disabled, but be sure SelectBurnInSamples is nonzero to guarantee
	// Expand is called.
	// Default is 100.
	MaxSpeculativeSamples int

	// RolloutsPerEpoch is a tuneable number of calls to Rollout per Search epoch.
	// This value should be high enough to amortize the cost of selecting a node
	// but not too high that it would take too much time away from other searches.
	// Default is 100.
	RolloutsPerEpoch int

	// ExplorationParameter is a tuneable parameter which weights the explore side of the
	// MAB policy.
	// Zero will use the default value of √2.
	ExplorationParameter float64
}

func (s *Search[E]) patchDefaults() {
	if s.Seed == 0 {
		s.Seed = time.Now().UnixNano()
	}
	if s.RolloutsPerEpoch == 0 {
		s.RolloutsPerEpoch = defaultRolloutsPerEpoch
	}
	if s.MaxSpeculativeSamples == 0 {
		s.MaxSpeculativeSamples = defaultMaxSelectSamples
	}
	if s.ExplorationParameter == 0 {
		s.ExplorationParameter = defaultExplorationParameter
	}
}

// Reset deletes the search continuation so the next call to Search starts from scratch.
func (s *Search[E]) Reset() {
	s.root = nil
}

func (c *Search[E]) Search(s SearchInterface[E], done <-chan struct{}) Stat[E] {
	c.patchDefaults()
	if c.root == nil {
		var sentinel E
		c.root = newEventLog(c, s, nil, sentinel, s.Log())
	}
	r := rand.New(rand.NewSource(c.Seed))
	root := c.root
	for {
		s.Root()
		node := root
		for {
			step, child, done := node.selectChild(r, c, s)
			if done {
				break
			}
			s.Apply(step)
			node = child
		}
		frontier := node
		frontierLog := s.Rollout()
		for i := 0; i < c.RolloutsPerEpoch-1; i++ {
			frontierLog = frontierLog.Merge(s.Rollout())
		}
		frontier.backprop(frontierLog, c.RolloutsPerEpoch)
		select {
		case <-done:
			return root.makeResult(r)
		default:
		}
	}
}

func uct(score float64, numRollouts, numParentRollouts int, explorationParameter float64) float64 {
	if numRollouts == 0 || numParentRollouts == 0 {
		return math.Inf(+1)
	}
	explore := explorationParameter * math.Sqrt(math.Log(float64(numParentRollouts))/float64(numRollouts))
	exploit := score / float64(numRollouts)
	return explore + exploit
}
