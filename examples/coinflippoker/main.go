package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/searchops"
)

type Pass struct{}

func (Pass) String() string { return "<pass>" }

type CoinFlip struct{}

func (CoinFlip) String() string { return "<coin-flip>" }

type Game struct {
	depth      int                         // depth determining player to move.
	limit      int                         // absolute limit of coin flip sums.
	pass       int                         // pass count. 1: Last turn of play. 2: game is over.
	playerSums model.TwoPlayerScalars[int] // player coin flip sums.
	r          *rand.Rand
}

func (g *Game) Root() { g.depth = 0; g.pass = 0; g.playerSums = [2]int{0, 0} }

func (g *Game) Expand(int) []mcts.FrontierAction {
	if g.pass == 2 {
		return nil
	}
	return []mcts.FrontierAction{{Action: CoinFlip{}}, {Action: Pass{}}}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *Game) Select(a mcts.Action) bool {
	switch a.(type) {
	case Pass:
		if g.pass++; g.pass > 2 {
			panic(fmt.Sprintf("Unexpected pass value: %d", g.pass))
		}
	case CoinFlip:
		player := g.depth & 1
		if abs(g.playerSums[player]) >= g.limit {
			return false
		}
		if g.Flip(player); g.pass == 1 {
			// This was the last turn of play. The game is over.
			g.pass = 2
		}
	}
	g.depth++ // Advance to next player.
	return true
}

func (g *Game) Flip(player int) {
	if g.r.Float32() < .5 {
		g.playerSums[player]++ // Heads is worth +1.
	} else {
		g.playerSums[player]-- // Tails is worth -1.
	}
}

var playerObjectives = [2]func(model.TwoPlayerScalars[int]) float64{
	model.MaximizePlayer1Scalars[int],
	model.MaximizePlayer2Scalars[int],
}

func (g Game) Score() mcts.Score[model.TwoPlayerScalars[int]] {
	player := g.depth & 1
	if g.depth > 0 {
		player = 1 - player
	}
	score := mcts.Score[model.TwoPlayerScalars[int]]{
		Counter:   model.TwoPlayerScalars[int]{},
		Objective: playerObjectives[player],
	}
	// Test loss conditions.
	for i := 0; i < 2; i++ {
		if g.playerSums[i] > g.limit {
			score.Counter[1-i]++
			return score
		}
	}
	if g.pass < 2 {
		return score
	}
	// Test end conditions if both players passed.
	s0, s1 := g.playerSums[0], g.playerSums[1]
	if s0 > s1 {
		score.Counter[0]++
	} else if s1 > s0 {
		score.Counter[1]++
	}
	return score
}

func main() {
	const limit = 3
	r := rand.New(rand.NewSource(1337))
	s := mcts.Search[model.TwoPlayerScalars[int]]{
		SearchInterface: &Game{limit: limit, r: r},
		AddCounters:     model.AddTwoPlayerScalars[int],
		Rand:            r,
	}
	for lastPrint := (time.Time{}); ; {
		s.Search()
		if time.Since(lastPrint) > time.Second {
			fmt.Println(searchops.FilterV[model.TwoPlayerScalars[int]](s.Tree,
				searchops.FilterNodePredicate(func(n mcts.Node[model.TwoPlayerScalars[int]]) bool { return n.NumRollouts > 0 }),
				searchops.HighestPriorityFilter[model.TwoPlayerScalars[int]](),
				searchops.AnyFilter[model.TwoPlayerScalars[int]](r)))
			lastPrint = time.Now()
		}
	}
}
