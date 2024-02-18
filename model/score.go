package model

import (
	"golang.org/x/exp/constraints"
)

type Scalar interface {
	constraints.Float | ~int | ~int64
}

func Add[T Scalar](c1, c2 T) T       { return c1 + c2 }
func Maximize[T Scalar](c T) float64 { return float64(c) }
func Minimize[T Scalar](c T) float64 { return float64(-c) }

type TwoPlayerScalars[T Scalar] [2]T

func AddTwoPlayerScalars[T Scalar](c1 *TwoPlayerScalars[T], c2 TwoPlayerScalars[T]) {
	c1[0] += c2[0]
	c1[1] += c2[1]
}

func TwoPlayerIndexByDepth(depth int) int {
	if depth == 0 || depth&1 == 1 {
		return 0
	}
	return 1
}

func MaximizePlayer1[T Scalar](c TwoPlayerScalars[T]) float64 { return float64(c[0] - c[1]) }
func MaximizePlayer2[T Scalar](c TwoPlayerScalars[T]) float64 { return float64(c[1] - c[0]) }

func MaximizeTwoPlayers[T Scalar]() [2]func(TwoPlayerScalars[T]) float64 {
	return [2]func(TwoPlayerScalars[T]) float64{MaximizePlayer1[T], MaximizePlayer2[T]}
}

func MinimizePlayer1[T Scalar](c TwoPlayerScalars[T]) float64 { return -MaximizePlayer1[T](c) }
func MinimizePlayer2[T Scalar](c TwoPlayerScalars[T]) float64 { return -MaximizePlayer2[T](c) }

func MinimizeTwoPlayers[T Scalar]() [2]func(TwoPlayerScalars[T]) float64 {
	return [2]func(TwoPlayerScalars[T]) float64{MinimizePlayer1[T], MinimizePlayer2[T]}
}

// ScorePlayer wraps a player index for multiplayer scorekeeping in zero-sum games.
//
// ScorePlayer is equivalent to ScoreWeights with Player's score weight of 1 and -1 elsewhere.
//
// ScorePlayer.Objective is an mcts.ObjectiveFunc.
type ScorePlayer[T Scalar] int

func (s ScorePlayer[T]) Objective(scores []T) float64 {
	switch len(scores) {
	case 2:
		return float64(scores[s] - scores[1-s])
	}
	score := float64(scores[s])
	n := len(scores)
	if n <= 1 {
		return score
	}
	w := 1 / float64(n-1)
	for _, x := range scores[:s] {
		score -= float64(x) * w
	}
	for _, x := range scores[s+1:] {
		score -= float64(x) * w
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
type ScoreWeights[T Scalar] []T

func (s ScoreWeights[T]) Objective(scores []T) float64 {
	score := float64(0)
	for i, x := range scores {
		score += float64(s[i] * x)
	}
	return score
}
