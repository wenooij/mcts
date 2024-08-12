package main

import (
	"fmt"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/searchops"
)

type countingAction int

const (
	sub countingAction = iota
	add
)

func (a countingAction) String() string {
	switch a {
	case sub:
		return "-"
	case add:
		return "+"
	default:
		return "?"
	}
}

type countingGame struct {
	Max         int
	n           int
	selectDepth int
}

func (c *countingGame) Root() {
	c.n = 50
	c.selectDepth = 0
}

func (c *countingGame) Select(a mcts.Action) bool {
	if c.selectDepth >= c.Max {
		return false
	}
	c.selectDepth++
	switch a.(countingAction) {
	case sub:
		c.n--
	case add:
		c.n++
	}
	return true
}

func (c *countingGame) Expand(int) []mcts.FrontierAction {
	if c.n <= 0 || c.n >= c.Max || c.selectDepth >= c.Max/2 {
		return nil
	}
	return []mcts.FrontierAction{
		{Action: sub},
		{Action: add},
	}
}

func (c countingGame) Score() mcts.Score[int] {
	score := mcts.Score[int]{Objective: model.Maximize[int]}
	if c.selectDepth >= c.Max/2 {
		score.Counter = c.n
	}
	score.Counter -= c.selectDepth
	return score
}

func (c countingGame) Hash() uint64 {
	return uint64(c.n)
}

func main() {
	const gameMax = 100

	g := &countingGame{Max: gameMax}
	s := &mcts.Search[int]{
		SearchInterface: model.MakeSearchInterface(g, mcts.CounterInterface[int]{}),
		NumEpisodes:     100,
	}

	start := time.Now()
	for i := 0; i < 10; i++ {
		s.Search()
	}
	fmt.Println("Search took", time.Since(start), " over ", s.NumEpisodes*10, "iterations")
	pv := searchops.PV(s, searchops.MaxDepthFilter[int](gameMax/2))
	fmt.Println(pv)
	fmt.Println("PV len = ", len(pv))
}
