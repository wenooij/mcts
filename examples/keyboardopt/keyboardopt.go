// Package keyboardopt runs an MCTS optimization search for the keyboard layout that
// minimizes travel distance from the home row for sequential block samples from
// targetsample.txt.
package main

import (
	_ "embed"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

//go:embed targetsample.txt
var targetSample string

const sampleLength = 100

type keySwapStep struct {
	p1 Pt
	p2 Pt
	ok bool
}

func (s keySwapStep) String() string {
	if s == (keySwapStep{}) {
		return "#"
	}
	return fmt.Sprintf("{%s:%s}", s.p1, s.p2)
}

type keyboardLog struct {
	travelDistance int
}

func (g *keyboardLog) Score() float64 {
	travelLoss := -float64(g.travelDistance)
	return travelLoss / sampleLength
}

func (g *keyboardLog) Merge(lg mcts.Log) mcts.Log {
	x := lg.(*keyboardLog)
	g.travelDistance += x.travelDistance
	return g
}

type keyboardNode struct {
	children map[keySwapStep]*keyboardNode
	step     keySwapStep

	depth  int
	layout Layout
}

func newRootKeyboardNode(r *rand.Rand) *keyboardNode {
	return &keyboardNode{
		children: make(map[keySwapStep]*keyboardNode, len(allKeys)),
		depth:    0,
		layout:   NewRandomLayout(r),
	}
}

func (n *keyboardNode) newChildKeyboardNode(step keySwapStep) *keyboardNode {
	depth := n.depth + 1
	child := &keyboardNode{
		children: make(map[keySwapStep]*keyboardNode, len(allKeys)-depth),
		step:     step,
		depth:    depth,
		layout:   n.layout.Clone(),
	}
	child.layout.Swap(step.p1, step.p2)
	return child
}

type keyboardSearch struct {
	r    *rand.Rand
	root *keyboardNode
	node *keyboardNode
}

func newKeyboardSearch(r *rand.Rand) *keyboardSearch {
	return &keyboardSearch{
		r:    r,
		root: newRootKeyboardNode(r),
	}
}

func (g *keyboardSearch) Select(step keySwapStep) {
	if child, ok := g.node.children[step]; ok {
		g.node = child
		return
	}
	child := g.node.newChildKeyboardNode(step)
	g.node.children[step] = child
	g.node = child
}

func (g *keyboardSearch) Root() {
	g.node = g.root
}

func (g *keyboardSearch) Log() mcts.Log {
	return &keyboardLog{}
}

func (g *keyboardSearch) Expand() []mcts.FrontierStep[keySwapStep] {
	if g.node.depth >= 10 {
		return nil
	}
	var steps []mcts.FrontierStep[keySwapStep]
	for i := 0; i < 10; i++ {
		p1, p2 := NewRandomValidPt(g.r), NewRandomValidPt(g.r)
		steps = append(steps, mcts.FrontierStep[keySwapStep]{Step: keySwapStep{p1, p2, true}})
	}
	return steps
}

func (g *keyboardSearch) Rollout() (mcts.Log, int) {
	i := rand.Intn(len(targetSample))
	end := i + sampleLength
	if end > len(targetSample) {
		end = len(targetSample)
	}
	sample := targetSample[i:end]

	score, _ := g.node.layout.Test(sample)
	return &keyboardLog{score}, 1
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := newKeyboardSearch(r)

	done := make(chan struct{})
	go func() {
		<-time.After(10 * time.Second)
		done <- struct{}{}
	}()

	opts := mcts.Search[keySwapStep]{
		ExploreFactor:   math.Pi,
		SearchInterface: s,
		Done:            done,
	}
	model.FitParams(&opts)
	fmt.Printf("Using c=%.4f\n---\n", opts.ExploreFactor)
	opts.Search()

	pv := opts.PV()
	fmt.Println(pv)

	layout := NewRandomLayout(r)
	for _, e := range pv {
		s := e.Step
		layout.Swap(s.p1, s.p2)
	}

	fmt.Println(layout)
}
