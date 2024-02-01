package mcts

import (
	"math/rand"
	"strconv"
)

// Copied from github.com/wenooij/model/dummy to use in tests.

type dummyScore float64

func (s dummyScore) Score() float64    { return float64(s) }
func (s dummyScore) Add(b Score) Score { return s + b.(dummyScore) }

type dummyAction int

func (s dummyAction) String() string { return strconv.FormatInt(int64(s), 10) }

type dummySearch struct {
	B    int
	Rand *rand.Rand
}

func (s dummySearch) Expand(n int) []FrontierAction {
	if n <= 0 {
		n = s.B
	}
	b := make([]FrontierAction, n)
	for i := range b {
		b[i] = FrontierAction{Action: dummyAction(i)}
	}
	return b
}

func (s dummySearch) Root()                 {}
func (s dummySearch) Select(Action)         {}
func (s dummySearch) Score() Score          { return dummyScore(s.Rand.NormFloat64()) }
func (s dummySearch) Rollout() (Score, int) { return s.Score(), 1 }
