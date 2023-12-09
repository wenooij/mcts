package mcts

import (
	"math"
)

const (
	explorationParameter = math.Sqrt2
	maxSelectSamples     = 10000
	rolloutsPerEpoch     = 10000
)

type Search struct {
	root *EventLog
}

func (c *Search) Search(s SearchInterface, done <-chan struct{}) Stat {
	if c.root == nil {
		c.root = newEventLog(nil, "")
	}
	root := c.root
	for {
		s.Root()
		node := root
		for {
			step, child := node.selectChild()
			if child == nil {
				if step = s.Expand(); step != "" {
					s.Apply(step)
					node = node.child(step)
				}
				break
			}
			s.Apply(step)
			node = node.child(step)
		}
		frontier := node
		var frontierLog Log
		for i := 0; i < rolloutsPerEpoch; i++ {
			frontierLog = s.Rollout(frontierLog)
		}
		frontier.backprop(frontierLog, rolloutsPerEpoch)
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
