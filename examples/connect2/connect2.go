package main

import (
	"fmt"
	"hash/maphash"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/searchops"
)

type stone byte

const (
	empty stone = iota
	white
	black
)

type action byte

func (a action) String() string { return string('0' + a) }

type connect2 struct {
	state      [4]stone
	depth      int
	objectives [2]func(model.TwoPlayerScalars[int]) float64
}

func (s connect2) winner() stone {
	last := s.state[0]
	for _, e := range s.state[1:] {
		if last == e {
			return e
		}
		last = e
	}
	return 0
}

func (s *connect2) nextStone() stone {
	return stone(1 + s.depth&1)
}

func (s *connect2) Select(a mcts.Action) bool {
	action := a.(action)
	s.state[action] = s.nextStone()
	s.depth++
	return true
}

func (s *connect2) Expand(int) []mcts.FrontierAction {
	if s.winner() != 0 {
		return nil
	}
	actions := make([]mcts.FrontierAction, 0, 4)
	for i, e := range s.state {
		if e == 0 {
			actions = append(actions, mcts.FrontierAction{Action: action(i)})
		}
	}
	return actions
}

func (s *connect2) Score() mcts.Score[model.TwoPlayerScalars[int]] {
	scores := mcts.Score[model.TwoPlayerScalars[int]]{
		Counter:   model.TwoPlayerScalars[int]{},
		Objective: s.objectives[model.TwoPlayerIndexByDepth(s.depth)],
	}
	switch s.winner() {
	case white:
		scores.Counter[0] = 1
	case black:
		scores.Counter[1] = 1
	}
	return scores
}

const initDepth = 0

func (s *connect2) Root() {
	s.depth = initDepth
	clear(s.state[:])
}

var seed = maphash.MakeSeed()

func (s *connect2) Hash() uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	h.WriteByte(byte(s.depth & 1))
	for _, b := range s.state {
		h.WriteByte(byte(b))
	}
	return h.Sum64()
}

func main() {
	cs := &connect2{objectives: model.MaximizeTwoPlayers[int]()}
	s := &mcts.Search[model.TwoPlayerScalars[int]]{
		SearchInterface: model.MakeSearchInterface[model.TwoPlayerScalars[int]](cs),
		AddCounters:     model.AddTwoPlayerScalars[int],
	}
	for lastTime := (time.Time{}); ; {
		if s.Search(); time.Since(lastTime) > time.Second {
			fmt.Println(searchops.PV(s))
			lastTime = time.Now()
		}
	}
}
