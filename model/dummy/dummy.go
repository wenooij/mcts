package dummy

import (
	"math/rand"
	"strconv"

	"github.com/wenooij/mcts"
)

type Step int

func (s Step) String() string { return strconv.FormatInt(int64(s), 10) }

type ScalarLog float64

func (x ScalarLog) Merge(lg mcts.Log) mcts.Log { x += lg.(ScalarLog); return x }
func (x ScalarLog) Score() float64             { return float64(x) }

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

func (s Search) Root()                    {}
func (s Search) Select(Step)              {}
func (s Search) Log() mcts.Log            { return ScalarLog(0) }
func (s Search) Rollout() (mcts.Log, int) { return ScalarLog(s.Rand.Float64()), 1 }
