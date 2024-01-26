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
	node  *tictactoeNode
	steps []mcts.FrontierStep[tictactoeStep]
}

func newSearchPlugin() *SearchPlugin {
	var n tictactoeNode
	n.Root()
	p := &SearchPlugin{
		node: &n,
	}
	p.steps = slices.Grow(p.steps, 9)
	return p
}

type tictactoeStep struct {
	cell byte
	turn byte
}

func (s tictactoeStep) String() string {
	if s == (tictactoeStep{}) {
		return "#"
	}
	return fmt.Sprintf("%c%c", '0'+s.cell, s.turn)
}

type tictactoeNode struct {
	depth    int
	terminal bool
	winner   byte
	state    [9]byte
	open     []byte
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
	n.open = append(n.open[:0],
		0, 1, 2,
		3, 4, 5,
		6, 7, 8,
	)
}

func (n *tictactoeNode) turn() byte {
	if n.depth&1 == 0 {
		return X
	}
	return O
}

func (n *tictactoeNode) computeTerminal() (winner byte, terminal bool) {
	if n.depth < 4 {
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
	case testRow(0) || testRow(3) || testRow(6) ||
		testCol(0) || testCol(1) || testCol(2) ||
		testDiag1() || testDiag2():
		n.terminal = true
		if n.turn() == X {
			return O, true
		}
		return X, true
	case len(n.open) == 0:
		n.terminal = true
		return 0, true
	default:
		return 0, false
	}
}

func (s *SearchPlugin) Root() {
	s.node.Root()
}

func (s *SearchPlugin) Expand() []mcts.FrontierStep[tictactoeStep] {
	if s.node.terminal {
		return nil
	}
	s.steps = s.steps[:0]
	for _, i := range s.node.open {
		s.steps = append(s.steps, mcts.FrontierStep[tictactoeStep]{
			Step: tictactoeStep{cell: i, turn: s.node.turn()}})
	}
	return s.steps
}

func (s *SearchPlugin) Select(step tictactoeStep) {
	n := s.node
	n.depth++
	idx := step.cell
	n.open = slices.DeleteFunc(n.open, func(i byte) bool { return i == idx })
	n.state[idx] = step.turn
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

	opts := mcts.Search[tictactoeStep]{
		SearchInterface: si,
		Done:            done,
	}
	model.FitParams(&opts)
	fmt.Printf("Using c=%.4f\n---\n", opts.ExploreFactor)
	opts.Search()

	fmt.Println(opts.PV())
}
