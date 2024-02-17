package model

import "golang.org/x/exp/constraints"

type Scalar interface {
	constraints.Float | int | int64
}

func AddScalar[T Scalar](c1, c2 T) T       { return c1 + c2 }
func MaximizeScalar[T Scalar](c T) float64 { return float64(c) }
func MinimizeScalar[T Scalar](c T) float64 { return float64(-c) }

type TwoPlayerScalars[T Scalar] [2]T

func AddTwoPlayerScalars[T Scalar](c1, c2 TwoPlayerScalars[T]) TwoPlayerScalars[T] {
	c1[0] += c2[0]
	c1[1] += c2[1]
	return c1
}

func MaximizePlayer1Scalars[T Scalar](c TwoPlayerScalars[T]) float64 { return float64(c[0] - c[1]) }
func MaximizePlayer2Scalars[T Scalar](c TwoPlayerScalars[T]) float64 { return float64(c[1] - c[0]) }

func MinimizePlayer1Scalars[T Scalar](c TwoPlayerScalars[T]) float64 { return float64(-c[0] + c[1]) }
func MinimizePlayer2Scalars[T Scalar](c TwoPlayerScalars[T]) float64 { return float64(-c[1] + c[0]) }

// ScorePlayer wraps a player index for multiplayer scorekeeping in zero-sum games.
//
// ScorePlayer is equivalent to ScoreWeights with Player's score weight of 1 and -1 elsewhere.
//
// ScorePlayer.Objective is an mcts.ObjectiveFunc.
type ScorePlayer int

func (s ScorePlayer) Objective(scores []float64) float64 {
	switch len(scores) {
	case 2:
		return scores[s] - scores[1-s]
	}
	score := scores[s]
	n := len(scores)
	if n <= 1 {
		return score
	}
	w := 1 / float64(n-1)
	for _, x := range scores[:s] {
		score -= x * w
	}
	for _, x := range scores[s+1:] {
		score -= x * w
	}
	return score
}

// ScoreWeights wraps a slice of float64s for multiplayer scorekeeping in generalized games
// where the payoff at a given position is a linear combination of player scores.
//
// Scores already implements weights for uniform zero-sum games. ScoreWeights is only
// for more complicated requirements. Note that situations that lead to positive or negative
// sums may impact the exploration tradeoff in search which assumes scores normalized to the
// interval [-1, +1]. Adjust ExploreFactor accordingly.
//
// ScoreWeights.Objective is an mcts.ObjectiveFunc.
type ScoreWeights []float64

func (s ScoreWeights) Objective(scores []float64) float64 {
	score := float64(0)
	for i, x := range scores {
		score += s[i] * x
	}
	return score
}
