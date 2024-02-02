// Package dummy implements a dummy search with random actions, infinite depth, and a score
// that quickly converges to the expected value of 1/2.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model/dummy"
)

func main() {
	B := flag.Int("b", 10, "Dummy branching factor")
	flag.Parse()

	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	flag.Parse()

	r := rand.New(rand.NewSource(*seed))

	done := make(chan struct{})
	go func() {
		<-time.After(1 * time.Second)
		done <- struct{}{}
	}()

	opts := mcts.Search{
		Rand:            r,
		Seed:            *seed,
		SearchInterface: dummy.Search{B: int(*B), Rand: r},
		ExploreFactor:   0.5,
	}
	for {
		opts.Search()
		select {
		case <-done:
			pv := opts.FilterV(
				mcts.PredicateFilter(func(e mcts.StatEntry) bool { return e.NumRollouts >= 1_000 }),
				mcts.AnyFilter(r))
			fmt.Println(pv)
			fmt.Println("---")
			fmt.Println(pv.Last().Score)
			return
		default:
		}
	}
}
