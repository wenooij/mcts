package mcts

type expandLimits[S Step] struct {
	*expandHeuristic

	depth int

	minExpandDepth           int
	maxSpeculativeExpansions int
}

func (e *expandLimits[S]) Init(parent *topo[S], h *expandHeuristic, minExpandDepth, maxSpeculativeExpansions int) {
	e.expandHeuristic = h
	if parent != nil {
		e.depth = parent.depth + 1
	}
	e.minExpandDepth = minExpandDepth
	e.maxSpeculativeExpansions = maxSpeculativeExpansions
}

func (e expandLimits[S]) Test() bool {
	return e.depth >= e.minExpandDepth && e.expandHeuristic.Samples() < e.maxSpeculativeExpansions
}
