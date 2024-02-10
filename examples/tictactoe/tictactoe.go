package main

import (
	"fmt"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

type mark byte

const (
	empty mark = iota
	X
	O
)

type SearchPlugin struct {
	node *tictactoeNode
}

func newSearchPlugin() *SearchPlugin {
	p := &SearchPlugin{
		node: new(tictactoeNode),
	}
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

func (n *tictactoeNode) player() int { return n.depth & 1 }
func (n *tictactoeNode) turn() mark  { return mark(1 + n.depth&1) }

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
	actions := make([]mcts.FrontierAction, 0, 9)
	for i, state := range s.node.state {
		if state != 0 {
			continue
		}
		actions = append(actions, mcts.FrontierAction{
			Action: tictactoeAction(i),
		})
	}
	return actions
}

func (s *SearchPlugin) Select(a mcts.Action) {
	n := s.node
	ta := a.(tictactoeAction)
	n.state[ta] = s.node.turn()
	n.depth++
}

func (s *SearchPlugin) Score() mcts.Score {
	// Depth penalty term rewards the earliest win.
	player := s.node.player()
	if s.node.depth > rootDepth {
		player = 1 - player
	}
	scores := model.Scores{Player: player, PlayerScores: make([]float64, 2)}
	d := float64(s.node.depth) / 1000
	switch s.node.computeWinner() {
	case X:
		scores.PlayerScores[0] = 1 - d
	case O:
		scores.PlayerScores[1] = 1 - d
	}
	return scores
}

func main() {
	si := newSearchPlugin()

	done := make(chan struct{})
	timer := time.After(60 * time.Second)
	go func() {
		<-timer
		done <- struct{}{}
	}()

	opts := mcts.Search{
		SearchInterface: si,
		NumEpisodes:     10000,
		ExploreFactor:   mcts.DefaultExploreFactor,
	}

	for lastPrint := time.Now(); ; {
		if opts.Search(); time.Since(lastPrint) >= time.Second {
			fmt.Println(opts.PV())
			lastPrint = time.Now()
		}
		select {
		case <-done:
			return
		default:
		}
	}
}
