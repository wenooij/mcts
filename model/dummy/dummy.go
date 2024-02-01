package dummy

import (
	"math/rand"
	"strconv"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

type Step int

func (s Step) String() string { return strconv.FormatInt(int64(s), 10) }

type Search struct {
	B    int
	Rand *rand.Rand
}

func (s Search) Expand(n int) []mcts.FrontierStep[Step] {
	if n <= 0 {
		n = s.B
	}
	b := make([]mcts.FrontierStep[Step], n)
	for i := range b {
		b[i] = mcts.FrontierStep[Step]{Step: Step(i)}
	}
	return b
}

func (s Search) Root()                      {}
func (s Search) Select(Step)                {}
func (s Search) Score() mcts.Score          { return model.Score(rand.NormFloat64()) }
func (s Search) Rollout() (mcts.Score, int) { return s.Score(), 1 }
