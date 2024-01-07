package main

import (
	"fmt"
	"math/rand"
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
	node *tictactoeNode
}

func newSearchPlugin() *SearchPlugin {
	var n tictactoeNode
	n.Root()
	return &SearchPlugin{
		node: &n,
	}
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

type tictactoeLog struct {
	turn   byte
	scoreX float64
	scoreO float64
}

func (e *tictactoeLog) Merge(lg mcts.Log) mcts.Log {
	t := lg.(*tictactoeLog)
	e.scoreX += t.scoreX
	e.scoreO += t.scoreO
	return e
}

func (e *tictactoeLog) Score() float64 {
	if e.turn == O {
		return float64(e.scoreX - e.scoreO)
	}
	return float64(e.scoreO - e.scoreX)
}

func (s *SearchPlugin) Root() {
	s.node.Root()
}

func (s *SearchPlugin) Expand() ([]tictactoeStep, bool) {
	n := s.node
	if n.terminal {
		return nil, true
	}
	i := n.open[rand.Intn(len(n.open))]
	return []tictactoeStep{{cell: i, turn: n.turn()}}, false
}

func (s *SearchPlugin) Apply(step tictactoeStep) {
	n := s.node
	n.depth++
	idx := step.cell
	n.open = slices.DeleteFunc(n.open, func(i byte) bool { return i == idx })
	n.state[idx] = step.turn
	n.winner, n.terminal = n.computeTerminal()
}

func (s *SearchPlugin) Log() mcts.Log {
	return &tictactoeLog{turn: s.node.turn()}
}

func (s *SearchPlugin) Rollout() (mcts.Log, int) {
	log := &tictactoeLog{turn: s.node.turn()}
	for s.forward(log) {
	}
	return log, 1
}

func (s *SearchPlugin) forward(log *tictactoeLog) bool {
	steps, _ := s.Expand()
	if len(steps) == 0 {
		switch s.node.winner {
		case X:
			log.scoreX++
		case O:
			log.scoreO++
		}
		return false
	}
	s.Apply(steps[0])
	return true
}

func main() {
	si := newSearchPlugin()

	done := make(chan struct{})
	timer := time.After(10 * time.Second)
	go func() {
		<-timer
		done <- struct{}{}
	}()

	opts := mcts.Search[tictactoeStep]{
		SearchInterface: si,
		Done:            done,
	}
	model.FitParams(&opts)
	fmt.Printf("Using c=%.4f\n---\n", opts.ExplorationParameter)
	opts.Search()

	fmt.Println(opts.PV())
}
