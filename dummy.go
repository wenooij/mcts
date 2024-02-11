package mcts

import (
	"math/rand"
	"strconv"
)

// Copied from github.com/wenooij/model/dummy to use in tests.

func maximizeObjective(scores []float64) float64 { return sumValues(scores) }

func sumValues(vs []float64) float64 {
	var sum float64
	for _, v := range vs {
		sum += v
	}
	return sum
}

type dummyAction int

func (s dummyAction) String() string { return strconv.FormatInt(int64(s), 10) }

type dummySearch struct {
	BranchFactor    int
	depth, MaxDepth int
	Rand            *rand.Rand
}

func (s dummySearch) Expand(n int) []FrontierAction {
	if s.MaxDepth > 0 && s.MaxDepth <= s.depth {
		return nil
	}
	b := make([]FrontierAction, s.BranchFactor)
	for i := range b {
		b[i] = FrontierAction{Action: dummyAction(i)}
	}
	return b
}

func (s *dummySearch) Root()         { s.depth = 0 }
func (s *dummySearch) Select(Action) { s.depth++ }
func (s dummySearch) Score() Score   { return Score{[]float64{s.Rand.NormFloat64()}, maximizeObjective} }
