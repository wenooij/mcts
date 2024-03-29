// Package model provides the model fitting utility FitParams.
package model

import (
	"fmt"
	"math"
	"slices"

	"github.com/wenooij/mcts"
)

const (
	numFitEpisodes = 10000
	numAnyVSamples = 255
	scoreSampleCap = 1024
)

type SummaryStats struct {
	N         int
	SampleN   int
	Min       float64
	Quartiles []float64
	Mean      float64
	Max       float64
	Stddev    float64
}

func (s *SummaryStats) ZScore(x float64) float64 {
	if s.SampleN == 0 {
		return x
	}
	return (x - s.Mean) / s.Stddev
}

func (s *SummaryStats) String() string {
	if s.SampleN == 0 {
		return "no summary"
	}
	q1, q2, q3 := math.NaN(), math.NaN(), math.NaN()
	if len(s.Quartiles) == 3 {
		q1 = s.Quartiles[0]
		q2 = s.Quartiles[1]
		q3 = s.Quartiles[2]
	}
	return fmt.Sprintf("min: %f [%f, %f, %f] mean: %f stddev: %f max: %f, over %d samples with reservoir size %d",
		s.Min,
		q1,
		q2,
		q3,
		s.Mean,
		s.Stddev,
		s.Max,
		s.N,
		s.SampleN,
	)
}

// FitParams tunes a search by computing stats from a smaller search
// and updating the ExploreFactor.
func Summarize[T mcts.Counter](s *mcts.Search[T]) SummaryStats {
	// Initialize the search for now, but leave the root as we found it.
	oldNumEpisodes := s.NumEpisodes
	defer func() { s.NumEpisodes = oldNumEpisodes }()
	s.NumEpisodes = numFitEpisodes
	if s.Init() {
		defer s.Reset()
	}
	s.Search()
	// Compute fit stats.
	var (
		minScore     = math.Inf(+1)
		maxScore     = math.Inf(-1)
		scoreSum     float64
		numSamples   int
		scoreSamples = make([]float64, 0, scoreSampleCap)
	)
	for i := 0; i < numAnyVSamples; i++ {
		// Sample score from random variations using reservoir sampling.
		var anyV []*mcts.Edge[T]
		root := s.RootEntry
		for n := root; n != nil; {
			child := (*n)[s.Rand.Intn(len(*n))]
			anyV = append(anyV, child)
			n = child.Dst
		}
		for _, e := range anyV {
			score := e.Score.Apply()
			if math.IsInf(score, 0) {
				// Inf scores cannot contribute to statistics.
				continue
			}
			if len(scoreSamples) < scoreSampleCap {
				scoreSamples = append(scoreSamples, score)
			} else if j := s.Rand.Intn(numSamples); j < scoreSampleCap {
				scoreSum -= scoreSamples[j]
				scoreSamples[j] = score
			}
			if score < minScore {
				minScore = score
			}
			if score > maxScore {
				maxScore = score
			}
			scoreSum += score
			numSamples++
		}
	}
	slices.Sort(scoreSamples) // Sort scores for quantiles.
	stats := SummaryStats{N: numSamples, SampleN: len(scoreSamples)}
	if stats.SampleN == 0 {
		return stats
	}
	stats.Min = minScore
	stats.Quartiles = []float64{
		scoreSamples[stats.SampleN/5],
		scoreSamples[stats.SampleN/2],
		scoreSamples[4*stats.SampleN/5],
	}
	stats.Mean = scoreSum / float64(stats.SampleN)
	stats.Max = maxScore

	// Compute stddev.
	var sdeSum float64
	for _, x := range scoreSamples {
		sdeSum += (x - stats.Mean) * (x - stats.Mean)
	}
	stats.Stddev = math.Sqrt(sdeSum / float64(stats.SampleN))
	return stats
}
