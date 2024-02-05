package model

import "github.com/wenooij/mcts"

// Score wraps a single float64 for single player scorekeeping.
//
// Satisfies mcts.Score.
type Score float64

func (s Score) Score() float64              { return float64(s) }
func (s Score) Add(b mcts.Score) mcts.Score { return s + b.(Score) }

// Scores wraps a slice of float64s for multiplayer scorekeeping in zero-sum games.
//
// Scores is equivalent to ScalarWeightedScores with Player's score weight of 1 and -1 elsewhere.
//
// Satisfies mcts.Score.
type Scores struct {
	Player       int
	PlayerScores []float64
}

func (s Scores) Add(b mcts.Score) mcts.Score {
	for i, x := range b.(Scores).PlayerScores {
		s.PlayerScores[i] += x
	}
	return s
}

func (s Scores) Score() float64 {
	score := s.PlayerScores[s.Player]
	n := len(s.PlayerScores)
	if n <= 1 {
		return score
	}
	w := 1 / float64(n-1)
	for _, x := range s.PlayerScores[:s.Player] {
		score -= x * w
	}
	for _, x := range s.PlayerScores[s.Player+1:] {
		score -= x * w
	}
	return score
}

// WeightedScores wraps a slice of float64s for multiplayer scorekeeping in generalized games
// where the payoff at a given position is a linear combination of player scores.
//
// Scores already implements weights for uniform zero-sum games. WeightedScores is only
// for more complicated requirements. Note that situations that lead to positive or negative
// sums may impact the exploration tradeoff in search which assumes scores normalized to the
// interval [-1, +1]. Adjust ExploreFactor accordingly.
//
// Satisfies mcts.Score.
type WeightedScores struct {
	Weights      []float64
	PlayerScores []float64
}

func (s WeightedScores) Add(b mcts.Score) mcts.Score {
	for i, x := range b.(WeightedScores).PlayerScores {
		s.PlayerScores[i] += x
	}
	return s
}
func (s WeightedScores) Score() float64 {
	score := float64(0)
	for i, x := range s.PlayerScores {
		score += s.Weights[i] * x
	}
	return score
}
