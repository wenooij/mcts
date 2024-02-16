package dummy

import (
	"math/rand"
	"strconv"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

type Action int

func (a Action) String() string { return strconv.FormatInt(int64(a), 10) }

type Search struct {
	B    int
	Rand *rand.Rand
}

func (s Search) Expand(n int) []mcts.FrontierAction {
	b := make([]mcts.FrontierAction, s.B)
	for i := range b {
		b[i] = mcts.FrontierAction{Action: Action(i)}
	}
	return b
}

func (s Search) Root()              {}
func (s Search) Select(mcts.Action) {}
func (s Search) Score() mcts.Score {
	return mcts.Score{Counters: []float64{rand.NormFloat64()}, Objective: model.MaximizeSum}
}
