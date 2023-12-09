package mcts

import (
	"math"
	"math/rand"
)

const (
	explorationParameter = math.Sqrt2
	maxSelectSamples     = 10000
	rolloutsPerEpoch     = 10000
)

type logTree struct {
	step     Step
	eventLog EventLog
	children map[Step]*logTree
}

func newLogTree(parent *logTree, step Step) *logTree {
	n := &logTree{
		step:     step,
		children: make(map[string]*logTree, 8),
	}
	if parent != nil {
		n.eventLog = EventLog{parent: &parent.eventLog}
	}
	return n
}

func (n *logTree) probablyExpandable() bool {
	log := n.eventLog
	samples, hits := log.NumSelectSamples, log.NumSelectHits
	return samples == 0 || hits == 0 ||
		samples < maxSelectSamples &&
			rand.Float64()*float64(samples) > float64(hits)
}

func (n *logTree) child(step Step) *logTree {
	n.eventLog.NumSelectSamples++
	child, ok := n.children[step]
	if ok {
		n.eventLog.NumSelectHits++
		return child
	}
	child = newLogTree(n, step)
	n.children[step] = child
	return child
}

func (n *logTree) selectChild() (Step, *logTree) {
	if n == nil || len(n.children) == 0 || n.probablyExpandable() {
		// We have no children or we think we can expand more.
		return "", nil
	}
	// Select an existing child to maximize policy.
	var (
		maxStep   Step
		maxChild  *logTree
		maxPolicy = -math.MaxFloat64
	)
	for k, v := range n.children {
		log := v.eventLog
		score, _ := log.Score()
		if policy := uct(score, log.NumRollouts, log.NumParentRollouts(), explorationParameter); maxPolicy < policy {
			maxStep = k
			maxChild = v
			maxPolicy = policy
		}
	}
	return maxStep, maxChild
}

func (n *logTree) bestChild() *logTree {
	if n == nil || len(n.children) == 0 {
		return nil
	}
	// Select an existing child to maximize score.
	var (
		maxChild *logTree
		maxScore = -math.MaxFloat64
	)
	for _, e := range n.children {
		log := e.eventLog
		score, _ := log.Score()
		score /= float64(log.NumRollouts)
		if maxScore < score {
			maxChild = e
			maxScore = score
		}
	}
	return maxChild
}

func (n *logTree) backprop(log Log, numRollouts int) {
	for e := &n.eventLog; e != nil; e = e.parent {
		e.NumRollouts += numRollouts
		if e.Log == nil {
			e.Log = log.Clone()
		} else {
			e.Log.Merge(log)
		}
	}
}

type Search struct {
	root *logTree
}

func (c *Search) Search(s SearchInterface, done <-chan struct{}) Stat {
	if c.root == nil {
		c.root = newLogTree(nil, "")
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
