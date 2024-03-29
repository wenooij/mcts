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

type Pass struct{}

func (Pass) String() string { return "<pass>" }

type CoinFlip struct{}

func (CoinFlip) String() string { return "<coin-flip>" }

type Game struct {
	depth      int                         // depth determining player to move.
	limit      int                         // absolute limit of coin flip sums.
	pass       int                         // pass count. 1: Last turn of play. 2: game is over.
	playerSums model.TwoPlayerScalars[int] // player coin flip sums.
	objectives [2]func(model.TwoPlayerScalars[int]) float64
	r          *rand.Rand
}

func newGame(r *rand.Rand, limit int) *Game {
	return &Game{
		limit:      limit,
		objectives: model.MaximizeTwoPlayers[int](),
		r:          r,
	}
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

func (g Game) Score() mcts.Score[model.TwoPlayerScalars[int]] {
	score := mcts.Score[model.TwoPlayerScalars[int]]{
		Counter:   model.TwoPlayerScalars[int]{},
		Objective: g.objectives[model.TwoPlayerIndexByDepth(g.depth)],
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

var seed = maphash.MakeSeed()

func (g *Game) Hash() uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	h.WriteByte(byte(g.depth & 1))
	h.WriteByte(byte(g.pass))
	h.WriteByte(byte(g.playerSums[0]))
	h.WriteByte(byte(g.playerSums[1]))
	return h.Sum64()
}

func main() {
	const limit = 3
	r := rand.New(rand.NewSource(1337))
	g := newGame(r, limit)
	s := mcts.Search[model.TwoPlayerScalars[int]]{
		SearchInterface: g,
		AddCounters:     model.AddTwoPlayerScalars[int],
		Rand:            r,
	}
	for lastPrint := (time.Time{}); ; {
		s.Search()
		if time.Since(lastPrint) > time.Second {
			fmt.Println(searchops.FilterV[model.TwoPlayerScalars[int]](s.RootEntry,
				searchops.EdgePredicate[model.TwoPlayerScalars[int]](func(n *mcts.Edge[model.TwoPlayerScalars[int]]) bool { return n.NumRollouts > 0 }).Filter,
				searchops.HighestPriorityFilter[model.TwoPlayerScalars[int]](),
				searchops.AnyFilter[model.TwoPlayerScalars[int]](r)))
			lastPrint = time.Now()
		}
	}
}
