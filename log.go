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

type EventLog[E Step] struct {
	parent   *EventLog[E]
	childSet map[E]*EventLog[E]
	children []*EventLog[E]

	Depth int
	Step  E
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

func newEventLog[E Step](c *Search[E], s SearchInterface[E], parent *EventLog[E], step E, log Log) *EventLog[E] {
	depth := 0
	if parent != nil {
		depth = parent.Depth + 1
	}
	return &EventLog[E]{
		parent:   parent,
		children: make([]*EventLog[E], 0, 8),
		childSet: make(map[E]*EventLog[E], 8),

		Depth: depth,
		Step:  step,
		Log:   log,

		MaxSelectSamples: c.MaxSpeculativeSamples,
	}
}

func (n *EventLog[E]) NumExpandHits() int {
	return n.NumExpandSamples - n.NumExpandMisses
}

func (n *EventLog[E]) expand(c *Search[E], s SearchInterface[E]) (step E, child *EventLog[E]) {
	n.NumExpandSamples++
	step = s.Expand()
	var sentinel E
	if terminal := step == sentinel; terminal {
		if !n.Terminal {
			n.Terminal = true
		} else {
			n.NumExpandMisses++
		}
		return step, nil
	}
	var ok bool
	if _, ok = n.childSet[step]; !ok {
		child = n.createChild(c, step, s)
	} else {
		// Add a sample miss when we have sampled it before.
		n.NumExpandMisses++
	}
	return step, child
}

func (n *EventLog[E]) burnIn(c *Search[E], s SearchInterface[E], runs int) {
	for i := 0; i < runs; i++ {
		n.expand(c, s)
	}
}

func (n *EventLog[E]) canStopHere(c *Search[E]) bool {
	return n.Depth >= c.MinSelectDepth
}

func (n *EventLog[E]) checkExpandHeuristic(rng *rand.Rand) bool {
	samples := n.NumExpandSamples
	if samples == 0 {
		return true
	}
	r := rng.Float64() * float64(samples)
	misses := float64(n.NumExpandMisses)
	return misses < r
}

func (n *EventLog[E]) selectChild(r *rand.Rand, c *Search[E], s SearchInterface[E]) (step E, child *EventLog[E], done bool) {
	if !n.BurnedIn {
		n.burnIn(c, s, c.SelectBurnInSamples)
		n.BurnedIn = true
		// Roll out from here, if possible.
		if n.canStopHere(c) {
			var sentinel E
			return sentinel, nil, true
		}
	}
	if n.NumExpandSamples < n.MaxSelectSamples+c.SelectBurnInSamples && n.checkExpandHeuristic(r) {
		// Try to further expand this node.
		n.expand(c, s)
		// Roll out from here, if possible.
		if n.canStopHere(c) {
			var sentinel E
			return sentinel, nil, true
		}
	}
	// Otherwise, select an existing child to maximize MAB policy.
	var (
		maxChild  *EventLog[E]
		maxPolicy = math.Inf(-1)
	)
	// Don't rely on the default map ordering.
	// Also required to make search repeatable.
	r.Shuffle(len(n.children), func(i, j int) { n.children[i], n.children[j] = n.children[j], n.children[i] })
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
		if math.IsInf(maxPolicy, +1) {
			break
		}
	}
	// Signal done if the selected step is a terminal.
	if terminal := maxChild == nil; terminal {
		var sentinel E
		return sentinel, nil, true
	}
	return maxChild.Step, maxChild, false
}

func (log *EventLog[E]) bestChild(r *rand.Rand) *EventLog[E] {
	if log == nil || len(log.children) == 0 {
		return nil
	}
	// Select an existing child to maximize score.
	var (
		maxChild *EventLog[E]
		maxScore = math.Inf(-1)
	)
	// Don't rely on the default map ordering.
	// Also required to make search repeatable.
	r.Shuffle(len(log.children), func(i, j int) { log.children[i], log.children[j] = log.children[j], log.children[i] })
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

func (n *EventLog[E]) backprop(log Log, numRollouts int) {
	for e := n; e != nil; e = e.parent {
		e.NumRollouts += numRollouts
		e.Log.Merge(log)
	}
}

func (n *EventLog[E]) createChild(c *Search[E], step E, s SearchInterface[E]) *EventLog[E] {
	child, ok := n.childSet[step]
	if ok {
		return child
	}
	child = newEventLog(c, s, n, step, s.Log())
	n.childSet[step] = child
	n.children = append(n.children, child)
	return child
}

func (e EventLog[E]) Score() (float64, bool) {
	if e.Log == nil {
		return 0, false
	}
	return e.Log.Score(), true
}

func (e EventLog[E]) NumParentRollouts() int {
	if e.parent == nil {
		return 0
	}
	return e.parent.NumRollouts
}
