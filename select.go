package mcts

import (
	"math"
	"math/rand"
)

type selector[S Step] struct {
	*topo[S]
	r *rand.Rand

	terminal bool

	explorationParameter float64
}

func (s *selector[S]) Init(n *topo[S], r *rand.Rand, exploarationParameter float64) {
	s.topo = n
	s.r = r
	s.explorationParameter = exploarationParameter
}

func (s *selector[S]) Select() (*topo[S], bool) {
	// Test the expand heuristic.
	if s.TryBurnIn(); s.expandHeuristic.Test() && s.expandLimits.Test() {
		// Heuristics suggest there may be new moves at this node
		// and expand limits do not prohibit expanding from this depth.
		return nil, false
	}
	// Otherwise, select an existing child to maximize MAB policy.
	var (
		maxChild  *topo[S]
		maxPolicy = math.Inf(-1)
	)
	for _, e := range s.children {
		if maxChild != nil {
			score, _ := e.Score()
			policy := uct(score, e.numRollouts, e.NumParentRollouts(), s.explorationParameter)
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
	if maxChild == nil {
		// There are no children for some reason.
		return nil, false
	}
	// Apply the step and return the max-policy child.
	return maxChild, true
}

func uct(score, numRollouts, numParentRollouts float64, explorationParameter float64) float64 {
	if numRollouts == 0 || numParentRollouts == 0 {
		return math.Inf(+1)
	}
	explore := explorationParameter * math.Sqrt(math.Log(float64(numParentRollouts))/float64(numRollouts))
	exploit := score / float64(numRollouts)
	return explore + exploit
}
