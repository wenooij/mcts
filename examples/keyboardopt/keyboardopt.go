// Package keyboardopt runs an MCTS optimization search for the keyboard layout that
// minimizes travel distance from the home row for sequential block samples from
// targetsample.txt.
package main

import (
	_ "embed"
	"fmt"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

//go:embed targetsample.txt
var targetSample string

const sampleLength = 100

type keySwapAction struct {
	p1 Pt
	p2 Pt
	ok bool
}

func (s keySwapAction) String() string {
	if s == (keySwapAction{}) {
		return "#"
	}
	return fmt.Sprintf("{%s:%s}", s.p1, s.p2)
}

type keyboardNode struct {
	children map[keySwapAction]*keyboardNode
	step     keySwapAction

	depth  int
	layout Layout
}

func newRootKeyboardNode(r *rand.Rand) *keyboardNode {
	return &keyboardNode{
		children: make(map[keySwapAction]*keyboardNode, len(allKeys)),
		depth:    0,
		layout:   NewRandomLayout(r),
	}
}

func (n *keyboardNode) newChildKeyboardNode(step keySwapAction) *keyboardNode {
	depth := n.depth + 1
	child := &keyboardNode{
		children: make(map[keySwapAction]*keyboardNode, len(allKeys)-depth),
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

func (g *keyboardSearch) Select(step keySwapAction) {
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

func (g *keyboardSearch) Expand(int) []mcts.FrontierAction[keySwapAction] {
	if g.node.depth > 13 {
		return nil
	}
	var actions []mcts.FrontierAction[keySwapAction]
	for i := 0; i < 20; i++ {
		p1, p2 := NewRandomValidPt(g.r), NewRandomValidPt(g.r)
		actions = append(actions, mcts.FrontierAction[keySwapAction]{Action: keySwapAction{p1, p2, true}})
	}
	return actions
}

func (g *keyboardSearch) Score() mcts.Score {
	i := rand.Intn(len(targetSample))
	end := i + sampleLength
	if end > len(targetSample) {
		end = len(targetSample)
	}
	sample := targetSample[i:end]
	score, _ := g.node.layout.Test(sample)
	return model.Score(-float64(score) / sampleLength)
}

func (g *keyboardSearch) Rollout() (mcts.Score, int) {
	return g.Score(), 1
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := newKeyboardSearch(r)

	done := make(chan struct{})
	go func() {
		<-time.After(30 * time.Second)
		done <- struct{}{}
	}()

	opts := mcts.Search[keySwapAction]{
		ExploreFactor:   40,
		SearchInterface: s,
	}
	for run := true; run; {
		opts.Search()
		select {
		case <-done:
			run = false
		default:
		}
	}

	pv := opts.PV()
	fmt.Println(pv)

	layout := NewRandomLayout(r)
	for _, e := range pv {
		s := e.Action
		layout.Swap(s.p1, s.p2)
	}

	fmt.Println(layout)
}
