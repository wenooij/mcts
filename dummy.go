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

type dummySearch struct{ Rand *rand.Rand }

func (s dummySearch) Expand(steps []FrontierStep[dummyStep]) (n int) {
	for i := 0; i < len(steps); i++ {
		steps[i] = FrontierStep[dummyStep]{
			Step:     dummyStep(i),
			Priority: s.Rand.Float64(),
		}
	}
	return len(steps)
}

func (s dummySearch) Root()               {}
func (s dummySearch) Select(dummyStep)    {}
func (s dummySearch) Log() Log            { return dummyScalarLog(0) }
func (s dummySearch) Rollout() (Log, int) { return dummyScalarLog(s.Rand.Float64()), 1 }
