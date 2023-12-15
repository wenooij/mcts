package mcts

import (
	"math"
	"math/rand"
)

// Log is used to keep track of the objective function value
// as well as aggregate events of interest.
type Log interface {
	// Merge in the provided EventLog.
	Merge(Log)
	// Score returns the objective function evaluation for this EventLog.
	Score() float64
}

type EventLog struct {
	parent   *EventLog
	children map[StepHash]*EventLog

	Depth int
	Step  Step
	Log   Log

	// Terminal is true when we have sampled an empty
	// step from this node at least once.
	Terminal bool
	// Explored is set when tuneable explore threshold is reached
	NumRollouts      int
	NumExpandMisses  int
	NumExpandSamples int
	MaxExpandSamples int
}

func newEventLog(c *Search, parent *EventLog, step Step, log Log) *EventLog {
	depth := 0
	if parent != nil {
		depth = parent.Depth + 1
	}
	return &EventLog{
		parent:   parent,
		children: make(map[StepHash]*EventLog, 8),

		Depth: depth,
		Step:  step,
		Log:   log,

		MaxExpandSamples: c.MaxExpandSamples,
	}
}

func (n *EventLog) NumExpandHits() int {
	return n.NumExpandSamples - n.NumExpandMisses
}

func (n *EventLog) expand(c *Search, s SearchInterface) (stepHash StepHash, child *EventLog) {
	n.NumExpandSamples++
	step := s.Expand()
	if terminal := step == nil; terminal {
		if !n.Terminal {
			n.Terminal = true
		} else {
			n.NumExpandMisses++
		}
		return 0, nil
	}
	stepHash = step.Hash()
	var ok bool
	if child, ok = n.children[stepHash]; !ok {
		child = n.child(c, step, s)
	} else {
		// Add a sample miss when we have sampled it before.
		n.NumExpandMisses++
	}
	return stepHash, child
}

func (n *EventLog) checkBurnIn(c *Search) bool {
	return n.NumExpandSamples <= c.ExtraExpandBurnInSamples
}

func (n *EventLog) canStopHere(c *Search) bool {
	return n.Depth >= c.MinSelectDepth
}

func (n *EventLog) checkExpandLimit() bool {
	return n.MaxExpandSamples <= 0 || n.NumExpandSamples < n.MaxExpandSamples
}

func (n *EventLog) checkExpandHeuristic() bool {
	samples, miss := n.NumExpandSamples, n.NumExpandMisses
	return samples == 0 || miss == 0 || rand.Float64()*float64(samples) < float64(miss)
}

func (n *EventLog) selectChild(c *Search, s SearchInterface) (step Step, child *EventLog, done bool) {
	if n.checkBurnIn(c) || n.checkExpandLimit() && n.checkExpandHeuristic() {
		// Try to further expand this node.
		// Either we have new node (or not yet reached the sample burn-in)
		// Or heuristics have told us to call Expand.
		_, child := n.expand(c, s)
		if done := child == nil || n.canStopHere(c); done {
			return nil, nil, true
		}
		if n.canStopHere(c) {
			return nil, nil, true
		}
		return child.Step, child, false
	}
	// Otherwise, select an existing child to maximize MAB policy.
	var (
		maxChild  *EventLog
		maxPolicy = -math.MaxFloat64
	)
	for _, e := range n.children {
		if maxChild != nil {
			score, _ := e.Score()
			policy := uct(score, e.NumRollouts, e.NumParentRollouts(), c.ExplorationParameter)
			if policy < maxPolicy {
				continue
			}
			maxPolicy = policy
		}
		maxChild = e
	}
	// Signal done if the selected step is a terminal.
	if terminal := maxChild == nil; terminal {
		return nil, nil, true
	}
	return maxChild.Step, maxChild, false
}

func (log *EventLog) bestChild() *EventLog {
	if log == nil || len(log.children) == 0 {
		return nil
	}
	// Select an existing child to maximize score.
	var (
		maxChild *EventLog
		maxScore = -math.MaxFloat64
	)
	for _, e := range log.children {
		score, _ := e.Score()
		score /= float64(e.NumRollouts)
		if maxScore < score {
			maxChild = e
			maxScore = score
		}
	}
	return maxChild
}

func (n *EventLog) backprop(log Log, numRollouts int) {
	for e := n; e != nil; e = e.parent {
		e.NumRollouts += numRollouts
		e.Log.Merge(log)
	}
}

func (n *EventLog) child(c *Search, step Step, s SearchInterface) *EventLog {
	stepHash := step.Hash()
	child, ok := n.children[stepHash]
	if ok {
		return child
	}
	child = newEventLog(c, n, step, s.Log())
	n.children[stepHash] = child
	return child
}

func (e EventLog) Score() (float64, bool) {
	if e.Log == nil {
		return 0, false
	}
	return e.Log.Score(), true
}

func (e EventLog) NumParentRollouts() int {
	if e.parent == nil {
		return 0
	}
	return e.parent.NumRollouts
}
