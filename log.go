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
	children map[Step]*EventLog

	Depth int
	Step  Step
	Log   Log

	// Terminal is true when we have sampled an empty
	// step from this node at least once.
	Terminal bool
	// Explored is set when tuneable explore threshold is reached
	NumRollouts      int
	NumExpandHits    int
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
		children: make(map[string]*EventLog, 8),

		Depth: depth,
		Step:  step,
		Log:   log,

		MaxExpandSamples: c.MaxExpandSamples,
	}
}

func (n *EventLog) expand(c *Search, s SearchInterface) (step Step, child *EventLog) {
	n.NumExpandSamples++
	step = s.Expand()
	if terminal := step == ""; terminal {
		if !n.Terminal {
			n.Terminal = true
		} else {
			n.NumExpandHits++
		}
		return "", nil
	}
	var ok bool
	if child, ok = n.children[step]; !ok {
		child = n.child(c, step, s)
	} else {
		// Add a sample hit when we have sampled it before.
		n.NumExpandHits++
	}
	return step, child
}

func (n *EventLog) sampleBurnIn(c *Search) bool {
	return n.NumExpandSamples <= c.ExtraExpandBurnInSamples
}

func (n *EventLog) canStopHere(c *Search) bool {
	return n.Depth >= c.MinSelectBurnInDepth
}

func (n *EventLog) stoppingHeuristic() bool {
	samples, hits := n.NumExpandSamples, n.NumExpandHits
	return samples > 0 && hits > 0 && rand.Float64()*float64(samples) < float64(hits)
}

func (n *EventLog) fullyExplored() bool {
	return n.MaxExpandSamples > 0 && n.NumExpandSamples < n.MaxExpandSamples
}

func (n *EventLog) selectChild(c *Search, s SearchInterface) (step Step, child *EventLog, done bool) {
	if n.sampleBurnIn(c) || !n.fullyExplored() {
		// Try to further expand this node.
		// Either we have new node (or not yet reached the sample burn-in)
		// Or heuristics have told us to call Expand.
		step, child = n.expand(c, s)
		// Signal done at this node if we are at a terminal or at the burn-in depth.
		terminal := step == ""
		if terminal || n.canStopHere(c) && n.stoppingHeuristic() {
			return "", nil, true
		}
		return step, child, false
	}
	// Otherwise, select an existing child to maximize MAB policy.
	var (
		maxStep   Step
		maxChild  *EventLog
		maxPolicy = -math.MaxFloat64
	)
	for k, e := range n.children {
		if maxStep != "" {
			score, _ := e.Score()
			policy := uct(score, e.NumRollouts, e.NumParentRollouts(), c.ExplorationParameter)
			if policy < maxPolicy {
				continue
			}
			maxPolicy = policy
		}
		maxStep = k
		maxChild = e
	}
	// Signal done if the selected step is a terminal.
	done = maxStep == ""
	return maxStep, maxChild, done
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
	child, ok := n.children[step]
	if ok {
		return child
	}
	child = newEventLog(c, n, step, s.Log())
	n.children[step] = child
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
