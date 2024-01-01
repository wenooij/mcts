package mcts

import "math/rand"

type expandHeuristic struct {
	hits, samples int

	maxSpeculativeExpansions int
}

func (h *expandHeuristic) Init(maxSpeculativeExpansions int) {
	h.maxSpeculativeExpansions = maxSpeculativeExpansions
}

func (h *expandHeuristic) Hit()  { h.Add(1, 1) }
func (h *expandHeuristic) Miss() { h.Add(0, 1) }

// Add a number of hits and misses.
//
// Hits are added when Expand yields a known Step.
// Misses are added when Expand yields a new Step.
// Samples are added whenever Expand is called.
func (h *expandHeuristic) Add(hits, samples int) {
	h.hits += hits
	h.samples += samples
}

// Samples returns the number of samples at this heuristic node.
func (h expandHeuristic) Samples() int { return h.samples }

// Test returns true if the explore heuristic is triggered, false otherwise.
//
// The test is based on a random chance proportional to the current miss rate.
// There is a short burn in period to establish the samples.
func (h *expandHeuristic) Test(r *rand.Rand) bool {
	if h.Samples() == 0 {
		// Always sample at least once.
		return true
	}
	if h.maxSpeculativeExpansions != 0 && h.Samples() >= h.maxSpeculativeExpansions {
		// Speculative samples disabled after this limit.
		return false
	}
	return r.Float64()*float64(h.samples) > float64(h.hits)
}