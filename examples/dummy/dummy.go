// Package dummy implements a dummy search with random actions, infinite depth, and a score
// that quickly converges to the expected value of 1/2.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/model/dummy"
	"github.com/wenooij/mcts/searchops"
)

func main() {
	B := flag.Int("b", 10, "Dummy branching factor")
	D := flag.Int("d", 10, "Dummy branching factor")
	flag.Parse()

	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	flag.Parse()

	r := rand.New(rand.NewSource(*seed))

	done := make(chan struct{})
	go func() {
		<-time.After(60 * time.Second)
		done <- struct{}{}
	}()

	dummy := &dummy.Search{BranchFactor: int(*B), MaxDepth: int(*D), Rand: r}
	s := &mcts.Search[float64]{
		Rand:            r,
		Seed:            *seed,
		SearchInterface: model.MakeSearchInterface(dummy, mcts.CounterInterface[float64]{}),
		ExploreFactor:   0.5,
	}
	for {
		s.Search()
		select {
		case <-done:
			pv := searchops.FilterV[float64](s.RootEntry,
				searchops.EdgePredicate[float64](func(n *mcts.Edge[float64]) bool { return n.NumRollouts >= 1_000 }).Filter,
				searchops.AnyFilter[float64](r))
			fmt.Println(pv)
			fmt.Println("---")
			fmt.Println(pv.Last().Score)
			return
		default:
		}
	}
}
