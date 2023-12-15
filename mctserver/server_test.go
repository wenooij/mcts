package mctserver

import (
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/wenooij/mcts"
)

func TestTicTacToeParallel(t *testing.T) {
	const workers = 1

	var wg sync.WaitGroup

	// simulate 2*simulatedCoreFactor cores
	const simulatedCoreFactor = 1
	const baseTime = 4 * time.Second
	searchTime := simulatedCoreFactor * baseTime

	mu := sync.Mutex{}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			si := NewPlugin()

			done := make(chan struct{})
			timer := time.After(searchTime)
			go func() {
				<-timer
				done <- struct{}{}
			}()

			c := new(mcts.Search)
			res := c.Search(si, done)

			mu.Lock()
			defer mu.Unlock()

			fmt.Println(res)
		}(i)
	}

	wg.Wait()
}

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

type tictactoeNode struct {
	depth    int
	terminal bool
	winner   byte
	state    [9]byte
	open     []byte
	children map[mcts.Step]*tictactoeNode
}

func newRoot() *tictactoeNode {
	return &tictactoeNode{
		state: [...]byte{
			0, 0, 0,
			0, 0, 0,
			0, 0, 0,
		},
		children: make(map[mcts.Step]*tictactoeNode, 9),
		open: []byte{
			0,
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
		},
	}
}

func newNode(parent *tictactoeNode, move mcts.Step) *tictactoeNode {
	d := parent.depth + 1
	n := &tictactoeNode{
		depth:    d,
		open:     make([]byte, len(parent.open)),
		children: make(map[mcts.Step]*tictactoeNode, 9-d),
	}
	copy(n.state[:], parent.state[:])
	idx := move[0] - '0'
	copy(n.open, parent.open)
	n.open = slices.DeleteFunc(n.open, func(i byte) bool { return i == idx })
	n.state[idx] = move[1]
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

func (s *SearchPlugin) Expand() mcts.Step {
	return s.node.expand()
}

func (n *tictactoeNode) expand() mcts.Step {
	if n.terminal {
		return ""
	}
	i := n.open[rand.Intn(len(n.open))]
	return string([]byte{'0' + i, n.turn()})
}

func (s *SearchPlugin) Apply(m mcts.Step) bool {
	s.node = s.node.apply(m)
	return true
}

func (n *tictactoeNode) apply(m mcts.Step) *tictactoeNode {
	child, ok := n.children[m]
	if ok {
		return child
	}
	child = newNode(n, m)
	n.children[m] = child
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
	if step == "" {
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

func NewPlugin() mcts.SearchInterface {
	return newSearchPlugin()
}
