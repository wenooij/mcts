// Package keyboardopt runs an MCTS optimization search for the keyboard layout that
// minimizes travel distance from the home row for sequential block samples from
// targetsample.txt.
package main

import (
	_ "embed"
	"fmt"
	"hash/maphash"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/searchops"
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

func (g *keyboardSearch) Select(a mcts.Action) bool {
	ksa := a.(keySwapAction)
	if child, ok := g.node.children[ksa]; ok {
		g.node = child
		return true
	}
	child := g.node.newChildKeyboardNode(ksa)
	g.node.children[ksa] = child
	g.node = child
	return true
}

func (g *keyboardSearch) Root() {
	g.node = g.root
}

func (g *keyboardSearch) Expand(int) []mcts.FrontierAction {
	if g.node.depth > 13 {
		return nil
	}
	var actions []mcts.FrontierAction
	for i := 0; i < 20; i++ {
		p1, p2 := NewRandomValidPt(g.r), NewRandomValidPt(g.r)
		actions = append(actions, mcts.FrontierAction{Action: keySwapAction{p1, p2, true}})
	}
	return actions
}

func (g *keyboardSearch) Score() mcts.Score[int] {
	i := rand.Intn(len(targetSample))
	end := i + sampleLength
	if end > len(targetSample) {
		end = len(targetSample)
	}
	sample := targetSample[i:end]
	score, _ := g.node.layout.Test(sample)
	return mcts.Score[int]{
		Counter:   score,
		Objective: model.Minimize[int],
	}
}

func (g *keyboardSearch) Rollout() (mcts.Score[int], float64) {
	return g.Score(), 1
}

var seed = maphash.MakeSeed()

func (g *keyboardSearch) Hash() uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	for _, r := range allKeys {
		pt := g.node.layout.Keys[r]
		h.WriteByte(pt.X)
		h.WriteByte(pt.Y)
	}
	return h.Sum64()
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ks := newKeyboardSearch(r)

	done := make(chan struct{})
	go func() {
		<-time.After(30 * time.Second)
		done <- struct{}{}
	}()

	s := &mcts.Search[int]{
		ExploreFactor:   40,
		SearchInterface: model.MakeSearchInterface[int](ks),
	}
	s.Root()
	for run := true; run; {
		s.Search()
		select {
		case <-done:
			run = false
		default:
		}
	}

	pv := searchops.PV(s)
	fmt.Println(pv)
	layout := ks.root.layout.Clone()
	for _, e := range pv {
		a := e.Action.(keySwapAction)
		layout.Swap(a.p1, a.p2)
	}
	fmt.Println(layout)
}
