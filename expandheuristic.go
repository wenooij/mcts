package mcts

import "math/rand"

const sampleBurnIn = 3

type expandHeuristic struct{ hits, samples int }

func (h *expandHeuristic) Hit()  { h.Add(1, 1) }
func (h *expandHeuristic) Miss() { h.Add(0, 1) }

// Add a number of hits and misses.
//
// Hits are added when Expand yields a new Step.
// Samples are added whenever Expand is called.
func (h *expandHeuristic) Add(hits, samples int) {
	h.hits += hits
	h.samples += samples
}

// Test returns true if the explore heuristic is triggered, false otherwise.
//
// The test is based on a random chance proportional to the current hit rate.
// There is a short burn in period to establish the samples.
func (h *expandHeuristic) Test(r *rand.Rand) bool {
	if h.samples < sampleBurnIn {
		// Always sample up to sampleBurnIn.
		return true
	}
	return r.Float64()*float64(h.samples) < float64(h.hits)
}
