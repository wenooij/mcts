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
	Min       float32
	Quartiles []float32
	Mean      float32
	Max       float32
	Stddev    float32
}

func (s *SummaryStats) ZScore(x float32) float32 {
	if s.SampleN == 0 {
		return x
	}
	return (x - s.Mean) / s.Stddev
}

func (s *SummaryStats) String() string {
	if s.SampleN == 0 {
		return "no summary"
	}
	q1, q2, q3 := float32(math.NaN()), float32(math.NaN()), float32(math.NaN())
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
func Summarize(s *mcts.Search) SummaryStats {
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
		minScore     = float32(math.Inf(+1))
		maxScore     = float32(math.Inf(-1))
		scoreSum     float32
		numSamples   int
		scoreSamples = make([]float32, 0, scoreSampleCap)
	)
	for i := 0; i < numAnyVSamples; i++ {
		// Sample score from random variations using reservoir sampling.
		for _, e := range s.AnyV() {
			if math.IsInf(float64(e.Score), 0) {
				// Inf scores cannot contribute to statistics.
				continue
			}
			if len(scoreSamples) < scoreSampleCap {
				scoreSamples = append(scoreSamples, e.Score)
			} else if j := s.Rand.Intn(numSamples); j < scoreSampleCap {
				scoreSum -= scoreSamples[j]
				scoreSamples[j] = e.Score
			}
			if e.Score < minScore {
				minScore = e.Score
			}
			if e.Score > maxScore {
				maxScore = e.Score
			}
			scoreSum += e.Score
			numSamples++
		}
	}
	slices.Sort(scoreSamples) // Sort scores for quantiles.
	stats := SummaryStats{N: numSamples, SampleN: len(scoreSamples)}
	if stats.SampleN == 0 {
		return stats
	}
	stats.Min = minScore
	stats.Quartiles = []float32{
		scoreSamples[stats.SampleN/5],
		scoreSamples[stats.SampleN/2],
		scoreSamples[4*stats.SampleN/5],
	}
	stats.Mean = scoreSum / float32(stats.SampleN)
	stats.Max = maxScore

	// Compute stddev.
	var sdeSum float32
	for _, x := range scoreSamples {
		sdeSum += (x - stats.Mean) * (x - stats.Mean)
	}
	stats.Stddev = float32(math.Sqrt(float64(sdeSum / float32(stats.SampleN))))
	return stats
}
