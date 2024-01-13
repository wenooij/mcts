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

type Search struct{ Rand *rand.Rand }

func (s Search) Expand(steps []mcts.FrontierStep[Step]) (n int) {
	for i := 0; i < len(steps); i++ {
		steps[i] = mcts.FrontierStep[Step]{
			Step:     Step(i),
			Priority: s.Rand.Float64(),
		}
	}
	return len(steps)
}

func (s Search) Root()                    {}
func (s Search) Select(Step)              {}
func (s Search) Log() mcts.Log            { return ScalarLog(0) }
func (s Search) Rollout() (mcts.Log, int) { return ScalarLog(s.Rand.Float64()), 1 }
