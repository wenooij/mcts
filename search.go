package mcts

import (
	"math"
)

const (
	defaultRolloutsPerEpoch     = 3
	defaultExplorationParameter = math.Sqrt2
)

// Search contains options used to run the MCTS Search.
// It also maintains a continuation which supports repeated calls to Search
// using the same search tree.
type Search struct {
	root *EventLog

	// MinSelectBurnInDepth is the minimum depth in which a rollout is allowed.
	// Before this depth we will rely entirely on Expand to give us nodes.
	// After this depth we may heuristically choose to Expand a nonterminal.
	// MinSelectBurnInDepth doesn't apply when Expand returns an empty step.
	// Default is 0.
	MinSelectBurnInDepth int

	// ExtraExpandBurnInSamples is the extra calls to Expand on a node before we
	// can start relying on sample heuristics.
	// Default is 0.
	ExtraExpandBurnInSamples int

	// MaxExpandSamples is the maximum number of calls to Expand on a node before we
	// start relying on the MAB policy during selection.
	// Default is 0 meaning unlimited samples.
	MaxExpandSamples int

	// RolloutsPerEpoch is the number of calls to SearchInterface's Rollout
	// per Select epoch. This should be set in accordance to the expensiveness of Rollout
	// to ensure exploration is done.
	// Default is 100.
	RolloutsPerEpoch int

	// ExplorationParameter is a tunable parameter which weights the explore side of the
	// MAB policy.
	// Zero will use the default value of √2.
	ExplorationParameter float64
}

func (s *Search) patchDefaults() {
	if s.RolloutsPerEpoch == 0 {
		s.RolloutsPerEpoch = defaultRolloutsPerEpoch
	}
	if s.ExplorationParameter == 0 {
		s.ExplorationParameter = defaultExplorationParameter
	}
}

// Reset deletes the search continuation so the next call to Search starts from scratch.
func (s *Search) Reset() {
	s.root = nil
}

func (c *Search) Search(s SearchInterface, done <-chan struct{}) Stat {
	c.patchDefaults()
	if c.root == nil {
		c.root = newEventLog(c, nil, "", s.Log())
	}
	root := c.root
	for {
		s.Root()
		node := root
		for {
			step, child, done := node.selectChild(c, s)
			if done {
				break
			}
			s.Apply(step)
			node = child
		}
		frontier := node
		frontierLog := s.Rollout()
		for i := 0; i < c.RolloutsPerEpoch-1; i++ {
			frontierLog.Merge(s.Rollout())
		}
		frontier.backprop(frontierLog, c.RolloutsPerEpoch)
		select {
		case <-done:
			return root.makeResult()
		default:
		}
	}
}

func uct(score float64, numRollouts, numParentRollouts int, explorationParameter float64) float64 {
	if numRollouts == 0 || numParentRollouts == 0 {
		return math.MaxFloat64
	}
	explore := explorationParameter * math.Sqrt(math.Log(float64(numParentRollouts))/float64(numRollouts))
	exploit := score / float64(numRollouts)
	return explore + exploit
}
