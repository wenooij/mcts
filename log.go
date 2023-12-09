package mcts

import (
	"math"
	"math/rand"
)

// Log is used to keep track of the objective function value
// as well as aggregate events of interest.
type Log interface {
	// Clone creates a copy of the Log.
	Clone() Log
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

	NumRollouts      int
	NumSelectHits    int
	NumSelectSamples int
}

func newEventLog(parent *EventLog, step Step) *EventLog {
	depth := 0
	if parent != nil {
		depth = parent.Depth + 1
	}
	return &EventLog{
		parent:   parent,
		children: make(map[string]*EventLog, 8),

		Depth: depth,
		Step:  step,
	}
}

func (log *EventLog) probablyExpandable() bool {
	samples, hits := log.NumSelectSamples, log.NumSelectHits
	return samples == 0 || hits == 0 ||
		samples < maxSelectSamples &&
			rand.Float64()*float64(samples) > float64(hits)
}

func (n *EventLog) selectChild() (Step, *EventLog) {
	if n == nil || len(n.children) == 0 || n.probablyExpandable() {
		// We have no children or we think we can expand more.
		return "", nil
	}
	// Select an existing child to maximize policy.
	var (
		maxStep   Step
		maxChild  *EventLog
		maxPolicy = -math.MaxFloat64
	)
	for k, e := range n.children {
		score, _ := e.Score()
		if policy := uct(score, e.NumRollouts, e.NumParentRollouts(), explorationParameter); maxPolicy < policy {
			maxStep = k
			maxChild = e
			maxPolicy = policy
		}
	}
	return maxStep, maxChild
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
		if e.Log == nil {
			e.Log = log.Clone()
		} else {
			e.Log.Merge(log)
		}
	}
}

func (n *EventLog) child(step Step) *EventLog {
	n.NumSelectSamples++
	child, ok := n.children[step]
	if ok {
		n.NumSelectHits++
		return child
	}
	child = newEventLog(n, step)
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
