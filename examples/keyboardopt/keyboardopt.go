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
	keysTyped      int
}

func (g *keyboardLog) Score() float64 {
	if g.keysTyped == 0 {
		return math.Inf(-1)
	}
	travelLoss := -float64(g.travelDistance)
	return travelLoss / float64(g.keysTyped)
}

func (g *keyboardLog) Merge(lg mcts.Log) {
	x := lg.(*keyboardLog)
	g.travelDistance += x.travelDistance
	g.keysTyped += x.keysTyped
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

func (g *keyboardSearch) Apply(step keySwapStep) {
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

func (g *keyboardSearch) Expand() keySwapStep {
	if g.node.depth >= 10 {
		return keySwapStep{}
	}
	p1, p2 := NewRandomValidPt(g.r), NewRandomValidPt(g.r)
	return keySwapStep{p1, p2, true}
}

func (g *keyboardSearch) Rollout() mcts.Log {
	i := rand.Intn(len(targetSample))
	end := i + sampleLength
	if end > len(targetSample) {
		end = len(targetSample)
	}
	sample := targetSample[i:end]

	score, hits := g.node.layout.Test(sample)
	return &keyboardLog{score, hits}
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := newKeyboardSearch(r)

	done := make(chan struct{})
	go func() {
		<-time.After(30 * time.Second)
		done <- struct{}{}
	}()

	opts := mcts.Search[keySwapStep]{
		MinSelectDepth:        5,
		SelectBurnInSamples:   5,
		MaxSpeculativeSamples: 5,
		RolloutsPerEpoch:      10,
		ExplorationParameter:  math.Pi,
	}
	res := opts.Search(s, done)

	fmt.Println(res)

	layout := NewRandomLayout(r)
	for leaf := res.PV; leaf != nil; leaf = leaf.PV {
		s := leaf.Step
		layout.Swap(s.p1, s.p2)
	}

	fmt.Println(layout)
}
