package main

import (
	"fmt"
	"hash/maphash"
	"math/rand"
	"slices"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/searchops"
)

type mark = byte

const (
	empty mark = iota
	X
	O
)

type SearchPlugin struct {
	node       *tictactoeNode
	actions    []mcts.FrontierAction
	objectives [2]func(model.TwoPlayerScalars[int64]) float64
	r          *rand.Rand
}

func newSearchPlugin() *SearchPlugin {
	p := &SearchPlugin{
		node:       new(tictactoeNode),
		objectives: model.MaximizeTwoPlayers[int64](),
		r:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	p.actions = slices.Grow(p.actions, 9)
	p.Root()
	return p
}

type tictactoeAction byte

func (s tictactoeAction) String() string {
	return fmt.Sprintf("%c", '0'+s)
}

type tictactoeNode struct {
	depth int
	state [9]mark
}

const rootDepth = 0

func (n *tictactoeNode) Root() {
	n.depth = rootDepth
	copy(n.state[:], []mark{
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
	})
}

func (n *tictactoeNode) turn() mark { return mark(1 + n.depth&1) }

func (n *tictactoeNode) computeWinner() (winner mark) {
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
		return s[0]
	case testRow(3), testCol(1), testDiag1(), testDiag2():
		return s[4]
	case testRow(6), testCol(2):
		return s[8]
	default:
		return 0
	}
}

func (s *SearchPlugin) Root() { s.node.Root() }

func (s *SearchPlugin) Expand(int) []mcts.FrontierAction {
	if s.node.computeWinner() != 0 {
		return nil
	}
	s.actions = s.actions[:0]
	for i, state := range s.node.state {
		if state != 0 {
			continue
		}
		s.actions = append(s.actions, mcts.FrontierAction{
			Action: tictactoeAction(i),
			Weight: rand.ExpFloat64(),
		})
	}
	return s.actions
}

func (s *SearchPlugin) Select(a mcts.Action) bool {
	n := s.node
	ta := a.(tictactoeAction)
	n.state[ta] = s.node.turn()
	n.depth++
	return true
}

func (s *SearchPlugin) Score() mcts.Score[model.TwoPlayerScalars[int64]] {
	// Depth penalty term rewards the earliest win.
	scores := mcts.Score[model.TwoPlayerScalars[int64]]{
		Counter:   model.TwoPlayerScalars[int64]{0, 0},
		Objective: s.objectives[model.TwoPlayerIndexByDepth(s.node.depth)],
	}
	// Applying a depth penalty enables MCTS to find the easrliest win, in theory.
	switch s.node.computeWinner() {
	case X:
		scores.Counter[0] = 100 - int64(s.node.depth)
	case O:
		scores.Counter[1] = 100 - int64(s.node.depth)
	}
	return scores
}

var seed = maphash.MakeSeed()

func (s *SearchPlugin) Hash() uint64 { return maphash.Bytes(seed, s.node.state[:]) }

func main() {
	si := newSearchPlugin()

	s := mcts.Search[model.TwoPlayerScalars[int64]]{
		SearchInterface: si,
		AddCounters:     model.AddTwoPlayerScalars[int64],
	}

	const epochs = 10000
	start := time.Now()
	for i := 0; i < epochs; i++ {
		s.Search()
	}
	fmt.Println("Search took", time.Since(start), "using", len(s.Table), "table entries and", s.NumEpisodes*epochs, "iterations")
	pv := searchops.PV(&s, searchops.MaxDepthFilter[model.TwoPlayerScalars[int64]](10))
	fmt.Println(pv)
}
