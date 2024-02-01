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
	if n <= 0 {
		n = s.B
	}
	b := make([]mcts.FrontierAction, n)
	for i := range b {
		b[i] = mcts.FrontierAction{Action: Action(i)}
	}
	return b
}

func (s Search) Root()                      {}
func (s Search) Select(mcts.Action)         {}
func (s Search) Score() mcts.Score          { return model.Score(rand.NormFloat64()) }
func (s Search) Rollout() (mcts.Score, int) { return s.Score(), 1 }
