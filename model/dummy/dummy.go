package dummy

import (
	"math/rand"
	"strconv"

	"github.com/wenooij/mcts"
)

type Action int

func (a Action) String() string { return strconv.FormatInt(int64(a), 10) }

type Search struct {
	BranchFactor    int
	depth, MaxDepth int
	Rand            *rand.Rand
	actions         []mcts.FrontierAction
}

func (s Search) Expand(n int) []mcts.FrontierAction {
	if s.depth >= s.MaxDepth {
		return nil
	}
	if len(s.actions) < s.BranchFactor {
		s.actions = make([]mcts.FrontierAction, s.BranchFactor)
		for i := range s.actions {
			s.actions[i] = mcts.FrontierAction{Action: Action(i)}
		}
	}
	return s.actions
}

func maximizeScalar(x float64) float64 { return x }

func (s Search) Root()                    {}
func (s *Search) Select(mcts.Action) bool { s.depth++; return true }
func (s Search) Score() mcts.Score[float64] {
	return mcts.Score[float64]{Counter: rand.NormFloat64(), Objective: maximizeScalar}
}
