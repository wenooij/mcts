package main

import (
	"fmt"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
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
	state [4]stone
	depth int
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

func (s *connect2) Select(a mcts.Action) {
	action := a.(action)
	s.state[action] = s.nextStone()
	s.depth++
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

func (s *connect2) Score() mcts.Score {
	player := int(s.nextStone() - 1)
	if s.depth > initDepth {
		player = 1 - player
	}
	scores := model.Scores{Player: player, PlayerScores: make([]float64, 2)}
	switch s.winner() {
	case white:
		scores.PlayerScores[0] = 1
	case black:
		scores.PlayerScores[1] = 1
	}
	return scores
}

const initDepth = 0

func (s *connect2) Root() {
	s.depth = initDepth
	copy(s.state[:], []stone{0, 0, 0, 0})
}

func main() {
	s := &mcts.Search{SearchInterface: &connect2{}}
	for lastTime := (time.Time{}); ; {
		if s.Search(); time.Since(lastTime) > time.Second {
			fmt.Println(s.PV())
			lastTime = time.Now()
		}
	}
}
