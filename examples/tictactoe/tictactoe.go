package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

const (
	X = byte('x')
	O = byte('o')
)

type SearchPlugin struct {
	node    *tictactoeNode
	actions []mcts.FrontierAction
}

func newSearchPlugin() *SearchPlugin {
	p := &SearchPlugin{
		node: new(tictactoeNode),
	}
	p.actions = slices.Grow(p.actions, 9)
	p.Root()
	return p
}

type tictactoeAction struct {
	cell byte
	turn byte
}

func (s tictactoeAction) String() string {
	if s == (tictactoeAction{}) {
		return "#"
	}
	return fmt.Sprintf("%c%c", '0'+s.cell, s.turn)
}

type tictactoeNode struct {
	depth    int
	terminal bool
	winner   byte
	state    [9]byte
}

func (n *tictactoeNode) Root() {
	n.depth = 0
	n.terminal = false
	n.winner = 0
	copy(n.state[:], []byte{
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
	})
}

func (n *tictactoeNode) turn() byte {
	if n.depth&1 == 0 {
		return X
	}
	return O
}

func (n *tictactoeNode) player() int {
	if n.depth&1 == 0 {
		return 0
	}
	return 1
}

func (n *tictactoeNode) computeTerminal() (winner byte, terminal bool) {
	if n.depth < 5 {
		return 0, false
	}
	s := n.state
	test := func(i, j, k int) bool {
		c := s[i]
		return c != 0 && c == s[j] && c == s[k]
	}
	testRow := func(i int) bool { return test(i, i+1, i+2) }
	testCol := func(i int) bool { return test(i, i+3, i+6) }
	testDiag1 := func() bool { return test(0, 4, 8) }
	testDiag2 := func() bool { return test(2, 4, 6) }
	switch {
	case testRow(0), testCol(0):
		return s[0], true
	case testRow(3), testCol(1), testDiag1(), testDiag2():
		return s[4], true
	case testRow(6), testCol(2):
		return s[8], true
	default:
		return 0, n.depth >= 9
	}
}

func (s *SearchPlugin) Root() {
	s.node.Root()
}

func (s *SearchPlugin) Expand(int) []mcts.FrontierAction {
	if s.node.terminal {
		return nil
	}
	s.actions = s.actions[:0]
	for i, state := range s.node.state {
		if state != 0 {
			continue
		}
		weight := 0.0
		a := tictactoeAction{cell: byte(i), turn: s.node.turn()}
		turn := s.node.turn()
		s.Select(a)
		if s.node.winner == turn {
			weight = 1000000
		}
		s.Unselect(a)
		s.actions = append(s.actions, mcts.FrontierAction{
			Action: a,
			Weight: weight,
		})
	}
	return s.actions
}

func (s *SearchPlugin) Select(a mcts.Action) {
	n := s.node
	n.depth++
	ta := a.(tictactoeAction)
	idx := ta.cell
	n.state[idx] = ta.turn
	n.winner, n.terminal = n.computeTerminal()
}

func (s *SearchPlugin) Unselect(step tictactoeAction) {
	n := s.node
	n.depth--
	idx := step.cell
	n.state[idx] = 0
	n.winner, n.terminal = n.computeTerminal()
}

func (s *SearchPlugin) Score() mcts.Score {
	scores := model.Scores{PlayerScores: make([]float64, 2)}
	if s.node.turn() == O {
		scores.Player = 1
	}
	if !s.node.terminal {
		return scores
	}
	switch s.node.winner {
	case X:
		scores.PlayerScores[0]++
	case O:
		scores.PlayerScores[1]++
	}
	return scores
}

func main() {
	si := newSearchPlugin()

	done := make(chan struct{})
	timer := time.After(1 * time.Second)
	go func() {
		<-timer
		done <- struct{}{}
	}()

	opts := mcts.Search{
		SearchInterface: si,
		NumEpisodes:     10000,
	}
	for {
		opts.Search()
		select {
		case <-done:
			fmt.Println(opts.PV())
			return
		default:
		}
	}
}
