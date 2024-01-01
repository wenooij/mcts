package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
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

func (n *nimState) Turn() int {
	return n.depth & 1
}

func (n *nimState) Result() (nimResult, bool) {
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
		return nimResult{}, false
	}
	turn := n.Turn()
	switch choices {
	case 0:
		res := nimResult{turn: turn}
		res.wins[turn]++
		return res, true
	case 1:
		res := nimResult{turn: turn}
		res.wins[1-turn]++
		return res, true
	}
	return nimResult{}, false
}

func (n *nimState) Apply(s nimStep) {
	n.piles[s.pile] -= nimPile(s.n)
}

func (n *nimState) Expand() (steps []nimStep, terminal bool) {
	for i, p := range n.piles {
		switch p {
		case 0:
		case 1:
			return []nimStep{{i, 1}}, false
		default:
			return []nimStep{{i, int(p)}, {i, int(p) - 1}}, false
		}
	}
	return nil, true
}

type nimResult struct {
	turn int
	wins [2]int
}

func (r nimResult) Merge(lg mcts.Log) mcts.Log {
	res := lg.(nimResult)
	r.wins[0] += res.wins[0]
	r.wins[1] += res.wins[1]
	return r
}

func (r nimResult) Score() float64 {
	if r.turn == 0 {
		return float64(r.wins[0] - r.wins[1])
	}
	return float64(r.wins[1] - r.wins[0])
}

func (n *nimState) Log() mcts.Log {
	return nimResult{turn: n.Turn()}
}

func (n *nimState) Rollout() (mcts.Log, int) {
	for {
		if r, ok := n.Result(); ok {
			return r, 1
		}
		s, _ := n.Expand()
		if len(s) == 0 {
			panic("no Steps but Result returned !ok")
		}
		n.Apply(s[n.r.Intn(len(s))])
	}
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

	opts := mcts.Search[nimStep]{}
	res := opts.Search(n, done)

	fmt.Println(res)
}
