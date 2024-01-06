package mcts

import "github.com/wenooij/heapordered"

func (h *node[S]) Hit()  { h.Add(1, 1) }
func (h *node[S]) Miss() { h.Add(0, 1) }

// Add a number of hits and misses.
//
// Hits are added when Expand yields a known Step.
// Misses are added when Expand yields a new Step.
// Samples are added whenever Expand is called.
func (h *node[S]) Add(hits, samples int) {
	h.hits += hits
	h.samples += samples
}

// Samples returns the number of samples at this heuristic node.
func (h *node[S]) Samples() int { return h.samples }

// Test returns true if the explore heuristic is triggered, false otherwise.
//
// The test is based on a random chance proportional to the current miss rate.
// There is a short burn in period to establish the samples.
func (h *node[S]) Test(s *Search[S], n *heapordered.Tree[*node[S]]) bool {
	if !h.burnedIn {
		h.burnIn(s, n)
	}
	if h.Samples() == 0 {
		// Always sample at least once.
		return true
	}
	if s.MaxSpeculativeExpansions != 0 && h.Samples() >= s.MaxSpeculativeExpansions {
		// Speculative samples disabled after this limit.
		return false
	}
	return s.Rand.Float64()*float64(h.samples) > float64(h.hits)
}

func (h *node[S]) burnIn(s *Search[S], n *heapordered.Tree[*node[S]]) bool {
	for i := 0; i < s.ExpandBurnInSamples; i++ {
		expand(s, n)
	}
	h.burnedIn = true
	return true
}
