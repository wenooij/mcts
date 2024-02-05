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

func (n *tictactoeNode) Root() {
	n.depth = 0
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
	scores := model.Scores{Player: 1 - s.node.player(), PlayerScores: make([]float64, 2)}
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
		InitRootScore:   func() mcts.Score { return model.Scores{PlayerScores: make([]float64, 2)} },
		ExploreFactor:   mcts.DefaultExploreFactor,
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		<-ch
		for _, e := range opts.PV() {
			fmt.Println(e)
		}
		fmt.Println("---")
		for _, a := range opts.RootActions() {
			fmt.Println(opts.Stat(a).Last())
		}
		fmt.Println("---")
		fmt.Println(opts.PV()[0].RawScore.(model.Scores))
		fmt.Println("---")

		var t gviz.Tree
		t.Add(opts.PV(), true)
		subtree := opts.Subtree(opts.PV().First().Action)
		for i := 0; i < 1000; i++ {
			t.Add(subtree.FilterV(
				mcts.MaxDepthFilter(3),
				mcts.AnyFilter(opts.Rand)), false)
		}
		f, err := os.OpenFile("tictactoe.dot", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
		if err != nil {
			panic(err)
		}
		if _, err := t.DOT(f); err != nil {
			panic(err)
		}
		f.Close()
		os.Exit(0)
	}()

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
