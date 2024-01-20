package mcts

import (
	"math/rand"
	"strconv"
)

// Copied from github.com/wenooij/model/dummy to use in tests.

type dummyStep int

func (s dummyStep) String() string { return strconv.FormatInt(int64(s), 10) }

type dummyScalarLog float64

func (x dummyScalarLog) Merge(lg Log) Log { x += lg.(dummyScalarLog); return x }
func (x dummyScalarLog) Score() float64   { return float64(x) }

type dummySearch struct {
	B    int
	Rand *rand.Rand
}

func (s dummySearch) Expand() []FrontierStep[dummyStep] {
	b := make([]FrontierStep[dummyStep], s.B)
	for i := range b {
		b[i] = FrontierStep[dummyStep]{
			Step:     dummyStep(i),
			Priority: s.Rand.Float64(),
		}
	}
	return b
}

func (s dummySearch) Root()               {}
func (s dummySearch) Select(dummyStep)    {}
func (s dummySearch) Log() Log            { return dummyScalarLog(0) }
func (s dummySearch) Rollout() (Log, int) { return dummyScalarLog(s.Rand.Float64()), 1 }
