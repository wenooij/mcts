package model

import (
	"math"
	"slices"

	"github.com/wenooij/mcts"
)

const (
	numFitEpochs    = 10000
	numAnyVSamples  = 255
	numScoreSamples = 1024
)

// FitParams tunes a search by computing stats from a smaller search
// and updating ExplorationParameter.
func FitParams[S mcts.Step](s *mcts.Search[S]) {
	// Initialize the search for now, but leave the root as we found it.
	if s.Init() {
		defer s.Reset()
	}
	for i := 0; i < numFitEpochs; i++ {
		s.SearchEpoch()
	}
	// Compute fit stats.
	numScores := 0
	scoreSamples := make([]float64, 0, numScoreSamples)
	for i := 0; i < numAnyVSamples; i++ {
		// Sample score from random variations using reservoir sampling.
		for _, e := range s.AnyV() {
			if math.IsInf(e.Score, -1) {
				// Nodes with no rollouts have the score -∞.
				continue
			}
			if len(scoreSamples) < cap(scoreSamples) {
				scoreSamples = append(scoreSamples, e.Score)
			} else if j := s.Rand.Intn(numScores); j < numScoreSamples {
				scoreSamples[j] = e.Score
			}
			numScores++
		}
	}
	// Fit ExplorationParameter to the absolute maximum of the first and third quantiles.
	slices.Sort(scoreSamples)
	exploreParam := math.Abs(scoreSamples[len(scoreSamples)/4])
	if score := math.Abs(scoreSamples[3*len(scoreSamples)/4]); score > exploreParam {
		exploreParam = score
	}
	s.ExplorationParameter = exploreParam
}
