package mcts

func (h *topo[S]) Hit()  { h.Add(1, 1) }
func (h *topo[S]) Miss() { h.Add(0, 1) }

// Add a number of hits and misses.
//
// Hits are added when Expand yields a known Step.
// Misses are added when Expand yields a new Step.
// Samples are added whenever Expand is called.
func (h *topo[S]) Add(hits, samples int) {
	h.hits += hits
	h.samples += samples
}

// Samples returns the number of samples at this heuristic node.
func (h *topo[S]) Samples() int { return h.samples }

// Test returns true if the explore heuristic is triggered, false otherwise.
//
// The test is based on a random chance proportional to the current miss rate.
// There is a short burn in period to establish the samples.
func (h *topo[S]) Test(s *Search[S]) bool {
	if !h.burnedIn {
		h.burnIn(s)
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

func (h *topo[S]) burnIn(s *Search[S]) bool {
	for i := 0; i < s.ExpandBurnInSamples; i++ {
		h.Expand(s)
	}
	h.burnedIn = true
	return true
}
