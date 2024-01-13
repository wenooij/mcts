// Package dummy implements a dummy search with random steps, infinite depth, and a score
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
)

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	flag.Parse()

	r := rand.New(rand.NewSource(*seed))

	done := make(chan struct{})
	go func() {
		<-time.After(1 * time.Second)
		done <- struct{}{}
	}()

	opts := mcts.Search[dummy.Step]{
		Rand:                     r,
		Seed:                     *seed,
		ExpandBurnInSamples:      10,
		ExpandBufferSize:         100,
		MaxSpeculativeExpansions: 10000,
		SearchInterface:          dummy.Search{Rand: r},
		Done:                     done,
	}
	model.FitParams(&opts)
	fmt.Printf("Using c=%.4f\n---\n", opts.ExplorationParameter)
	opts.Search()

	pv := opts.FilterV(
		mcts.PredicateFilter(func(e mcts.StatEntry[dummy.Step]) bool { return e.NumRollouts >= 1_000 }),
		mcts.AnyFilter[dummy.Step](r))
	fmt.Println(pv)
	fmt.Println("---")
	fmt.Println(pv.Leaf().Score)
}
