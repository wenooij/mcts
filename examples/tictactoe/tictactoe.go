package main

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/wenooij/mcts"
)

const (
	X = byte('x')
	O = byte('o')
)

type SearchPlugin struct {
	root *tictactoeNode
	node *tictactoeNode
}

func newSearchPlugin() *SearchPlugin {
	root := newRoot()
	return &SearchPlugin{
		root: root,
		node: root,
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
	children map[tictactoeStep]*tictactoeNode
}

func newRoot() *tictactoeNode {
	return &tictactoeNode{
		state: [...]byte{
			0, 0, 0,
			0, 0, 0,
			0, 0, 0,
		},
		children: make(map[tictactoeStep]*tictactoeNode, 9),
		open: []byte{
			0, 1, 2,
			3, 4, 5,
			6, 7, 8,
		},
	}
}

func newNode(parent *tictactoeNode, step tictactoeStep) *tictactoeNode {
	d := parent.depth + 1
	n := &tictactoeNode{
		depth:    d,
		open:     make([]byte, len(parent.open)),
		children: make(map[tictactoeStep]*tictactoeNode, 9-d),
	}
	copy(n.state[:], parent.state[:])
	idx := step.cell
	copy(n.open, parent.open)
	n.open = slices.DeleteFunc(n.open, func(i byte) bool { return i == idx })
	n.state[idx] = step.turn
	n.winner, n.terminal = n.computeTerminal()
	return n
}

func (n *tictactoeNode) turn() byte {
	if n.depth&1 == 0 {
		return X
	}
	return O
}

func (n *tictactoeNode) computeTerminal() (winner byte, terminal bool) {
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

func (e *tictactoeLog) Merge(lg mcts.Log) {
	t := lg.(*tictactoeLog)
	e.scoreX += t.scoreX
	e.scoreO += t.scoreO
}

func (e *tictactoeLog) Score() float64 {
	if e.turn == O {
		return float64(e.scoreX - e.scoreO)
	}
	return float64(e.scoreO - e.scoreX)
}

func (s *SearchPlugin) Root() {
	s.node = s.root
}

func (s *SearchPlugin) Expand() tictactoeStep {
	return s.node.expand()
}

func (n *tictactoeNode) expand() tictactoeStep {
	if n.terminal {
		return tictactoeStep{}
	}
	i := n.open[rand.Intn(len(n.open))]
	return tictactoeStep{cell: i, turn: n.turn()}
}

func (s *SearchPlugin) Apply(step tictactoeStep) {
	s.node = s.node.apply(step)
}

func (n *tictactoeNode) apply(step tictactoeStep) *tictactoeNode {
	child, ok := n.children[step]
	if ok {
		return child
	}
	child = newNode(n, step)
	n.children[step] = child
	return child
}

func (s *SearchPlugin) Log() mcts.Log {
	return &tictactoeLog{turn: s.node.turn()}
}

func (s *SearchPlugin) Rollout() mcts.Log {
	frontier := s.node
	defer func() { s.node = frontier }()
	log := &tictactoeLog{turn: s.node.turn()}
	for s.forward(log) {
	}
	return log
}

func (s *SearchPlugin) forward(log *tictactoeLog) bool {
	step := s.node.expand()
	if step == (tictactoeStep{}) {
		switch s.node.winner {
		case X:
			log.scoreX++
		case O:
			log.scoreO++
		}
		return false
	}
	s.node = s.node.apply(step)
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
		MinSelectDepth:        0,
		SelectBurnInSamples:   100,
		MaxSpeculativeSamples: 20,
		RolloutsPerEpoch:      10000,
		ExplorationParameter:  math.Sqrt2 / 10,
	}
	res := opts.Search(si, done)

	fmt.Println(res)
}
