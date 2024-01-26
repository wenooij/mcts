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

func (s Search) Expand() []mcts.FrontierStep[Step] {
	b := make([]mcts.FrontierStep[Step], s.B)
	for i := range b {
		b[i] = mcts.FrontierStep[Step]{
			Step:     Step(i),
			Priority: s.Rand.Float64(),
		}
	}
	return b
}

func (s Search) Root()                      {}
func (s Search) Select(Step)                {}
func (s Search) Score() mcts.Score          { return model.Score(rand.NormFloat64()) }
func (s Search) Rollout() (mcts.Score, int) { return s.Score(), 1 }
