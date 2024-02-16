package model

func MinimizeSum(scores []float64) float64 { return -sumValues(scores) }
func MaximizeSum(scores []float64) float64 { return sumValues(scores) }

func sumValues(vs []float64) float64 {
	switch len(vs) {
	case 1:
		return vs[0]
	case 2:
		return vs[0] + vs[1]
	}
	var sum float64
	for _, v := range vs {
		sum += v
	}
	return sum
}

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
