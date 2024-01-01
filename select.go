package mcts

import (
	"math"
	"math/rand"
)

type selector[S Step] struct {
	*topo[S]

	terminal bool

	explorationParameter float64
}

func (s *selector[S]) Init(n *topo[S], exploarationParameter float64) {
	s.topo = n
	s.explorationParameter = exploarationParameter
}

func (s *selector[S]) Select(r *rand.Rand) (*topo[S], bool) {
	// Test the expand heuristic.
	if s.TryBurnIn() || s.expandHeuristic.Test(r) {
		// Heuristics suggest there may be new moves at this node
		// and expand limits do not prohibit expanding from this depth.
		return nil, false
	}
	// Otherwise, select an existing child to maximize MAB policy.
	child := s.childByPolicy(r)
	return child, child != nil
}

// childByPolicy selects an existing child to maximize MAB policy or returns nil
func (t *topo[S]) childByPolicy(r *rand.Rand) *topo[S] {
	var (
		maxChild  *topo[S]
		maxPolicy = math.Inf(-1)
	)
	for _, e := range t.children {
		if maxChild != nil {
			score, _ := e.Score()
			policy := uct(score, e.numRollouts, e.NumParentRollouts(), t.explorationParameter)
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
		return nil
	}
	return maxChild
}

func uct(score, numRollouts, numParentRollouts float64, explorationParameter float64) float64 {
	if numRollouts == 0 || numParentRollouts == 0 {
		return math.Inf(+1)
	}
	explore := explorationParameter * math.Sqrt(math.Log(float64(numParentRollouts))/float64(numRollouts))
	exploit := score / float64(numRollouts)
	return explore + exploit
}
