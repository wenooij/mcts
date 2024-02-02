package mcts

import (
	"math/rand"
	"strconv"
)

// Copied from github.com/wenooij/model/dummy to use in tests.

type dummyScore float32

func (s dummyScore) Score() float32    { return float32(s) }
func (s dummyScore) Add(b Score) Score { return s + b.(dummyScore) }

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

func (s *dummySearch) Root()                { s.depth = 0 }
func (s *dummySearch) Select(Action)        { s.depth++ }
func (s dummySearch) Score() Score          { return dummyScore(s.Rand.NormFloat64()) }
func (s dummySearch) Rollout() (Score, int) { return s.Score(), 1 }
