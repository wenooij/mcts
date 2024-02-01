package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

type nimStep struct {
	pile int
	n    int
}

func (s nimStep) String() string {
	if s.n == 0 {
		return "#"
	}
	return fmt.Sprintf("%d(¡%d)", s.pile, s.n)
}

type nimState struct {
	N     int
	r     *rand.Rand
	piles []nimPile
	depth int
}

func (n *nimState) Root() {
	if n.N == 0 {
		n.N = 4
	}
	// 0    ¡
	// 1   ¡¡¡
	// 2  ¡¡¡¡¡
	// 3 ¡¡¡¡¡¡¡
	n.piles = n.piles[:0]
	for i := 0; i < n.N; i++ {
		n.piles = append(n.piles, nimPile(2*i+1))
	}
}

func (n *nimState) Player() int {
	return n.depth & 1
}

func (n *nimState) Choices() int {
	var choices int
	for _, p := range n.piles {
		if p == 0 {
			continue
		}
		if p == 1 {
			if choices++; choices <= 1 {
				continue
			}
		}
		break
	}
	return choices
}

func (n *nimState) Score() mcts.Score {
	player := n.Player()
	scores := model.Scores{Player: player, PlayerScores: make([]float64, 2)}
	switch n.Choices() {
	case 0:
		scores.PlayerScores[player]++
		return scores
	case 1:
		scores.PlayerScores[1-player]++
		return scores
	}
	return scores
}

func (n *nimState) Select(s nimStep) {
	n.piles[s.pile] -= nimPile(s.n)
}

func (s *nimState) Expand(int) []mcts.FrontierStep[nimStep] {
	var steps []mcts.FrontierStep[nimStep]
	for i, p := range s.piles {
		switch p {
		case 0:
		case 1:
			steps = append(steps, mcts.FrontierStep[nimStep]{Step: nimStep{i, 1}})
		default:
			steps = append(steps, mcts.FrontierStep[nimStep]{Step: nimStep{i, int(p)}},
				mcts.FrontierStep[nimStep]{Step: nimStep{i, int(p) - 1}})
		}
	}
	return steps
}

type nimPile int

func main() {
	r := rand.New(rand.NewSource(1337))
	n := &nimState{N: 4, r: r}
	n.Root()

	done := make(chan struct{})
	go func() {
		time.Sleep(5 * time.Second)
		done <- struct{}{}
	}()

	opts := mcts.Search[nimStep]{SearchInterface: n}
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
