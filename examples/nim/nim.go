package main

import (
	"fmt"
	"hash/maphash"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/searchops"
)

type nimPile int

type nimAction struct {
	pile int
	n    int
}

func (s nimAction) String() string {
	if s.n == 0 {
		return "#"
	}
	return fmt.Sprintf("%d(¡%d)", s.pile, s.n)
}

type nimState struct {
	N          int
	r          *rand.Rand
	piles      []nimPile
	depth      int
	objectives [2]func([2]int) float64
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

func (n *nimState) Score() mcts.Score[[2]int] {
	player := n.Player()
	scores := mcts.Score[[2]int]{
		Counter:   [2]int{},
		Objective: n.objectives[model.TwoPlayerIndexByDepth(n.depth)],
	}
	switch n.Choices() {
	case 0:
		scores.Counter[player]++
		return scores
	case 1:
		scores.Counter[1-player]++
		return scores
	}
	return scores
}

func (n *nimState) Select(a mcts.Action) bool {
	na := a.(nimAction)
	n.piles[na.pile] -= nimPile(na.n)
	return true
}

func (s *nimState) Expand(int) []mcts.FrontierAction {
	var actions []mcts.FrontierAction
	for i, p := range s.piles {
		switch p {
		case 0:
		case 1:
			actions = append(actions, mcts.FrontierAction{Action: nimAction{i, 1}})
		default:
			actions = append(actions, mcts.FrontierAction{Action: nimAction{i, int(p)}},
				mcts.FrontierAction{Action: nimAction{i, int(p) - 1}})
		}
	}
	return actions
}

var seed = maphash.MakeSeed()

func (s *nimState) Hash() uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	h.WriteByte(byte(s.depth & 1))
	for _, p := range s.piles {
		h.WriteByte(byte(p))
	}
	return h.Sum64()
}

func main() {
	r := rand.New(rand.NewSource(1337))
	n := &nimState{N: 4, r: r, objectives: model.MaximizeTwoPlayers[int]()}
	n.Root()

	s := &mcts.Search[[2]int]{
		SearchInterface: model.MakeSearchInterface(n, model.TwoPlayerScalarsInterface[int]()),
	}
	for lastTime := (time.Time{}); ; {
		if s.Search(); time.Since(lastTime) > time.Second {
			fmt.Println(searchops.PV(s))
			lastTime = time.Now()
		}
	}
}
