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
	// BurnIn has run on this node.
	BurnedIn bool
	// Explored is set when tuneable explore threshold is reached
	NumRollouts      int
	NumExpandMisses  int
	NumExpandSamples int
	MaxSelectSamples int
}

func newEventLog(c *Search, s SearchInterface, parent *EventLog, step Step, log Log) *EventLog {
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

		MaxSelectSamples: c.SelectSamples,
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
		child = n.createChild(c, step, s)
	} else {
		// Add a sample miss when we have sampled it before.
		n.NumExpandMisses++
	}
	return stepHash, child
}

func (n *EventLog) burnIn(c *Search, s SearchInterface, runs int) {
	for i := 0; i < runs; i++ {
		step := s.Expand()
		n.createChild(c, step, s)
	}
}

func (n *EventLog) canStopHere(c *Search) bool {
	return n.Depth >= c.MinSelectDepth
}

func (n *EventLog) checkExpandHeuristic() bool {
	samples, miss := n.NumExpandSamples, n.NumExpandMisses
	return samples == 0 || miss == 0 || rand.Float64()*float64(samples) < float64(miss)
}

func (n *EventLog) selectChild(c *Search, s SearchInterface) (step Step, child *EventLog, done bool) {
	if n.BurnedIn {
		n.burnIn(c, s, 1+c.SelectBurnInSamples)
		n.BurnedIn = true
	}
	if n.NumExpandSamples < n.MaxSelectSamples && n.checkExpandHeuristic() {
		// Try to further expand this node.
		// Either we have new node (or not yet reached the sample burn-in)
		// Or heuristics have told us to call Expand.
		_, child := n.expand(c, s)
		if done := child == nil || n.canStopHere(c); done {
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

func (n *EventLog) createChild(c *Search, step Step, s SearchInterface) *EventLog {
	stepHash := step.Hash()
	child, ok := n.children[stepHash]
	if ok {
		return child
	}
	child = newEventLog(c, s, n, step, s.Log())
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
