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

const sampleLength = 1000

type keyboardLog struct {
	travelDistance int
	keysTyped      int
}

func (g *keyboardLog) Score() float64 {
	if g.keysTyped == 0 {
		return -math.MaxFloat64
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
	children map[string]*keyboardNode
	step     string

	depth  int
	layout Layout
}

func newRootKeyboardNode(r *rand.Rand) *keyboardNode {
	return &keyboardNode{
		children: make(map[string]*keyboardNode, len(allKeys)),
		depth:    0,
		layout:   NewRandomLayout(r),
	}
}

func parseStep(step string) (Pt, Pt) {
	var x1, y1 int
	var x2, y2 int
	if n, err := fmt.Sscan(step, &x1, &y1, &x2, &y2); err != nil {
		panic(fmt.Errorf("parseStep: failed for step %q: %v at %d", step, err, n))
	}
	p1 := Pt{x1, y1}
	p2 := Pt{x2, y2}
	return p1, p2
}

func (n *keyboardNode) newChildKeyboardNode(step string) *keyboardNode {
	depth := n.depth + 1
	child := &keyboardNode{
		children: make(map[string]*keyboardNode, len(allKeys)-depth),
		step:     step,
		depth:    depth,
		layout:   n.layout.Clone(),
	}
	p1, p2 := parseStep(step)
	child.layout.Swap(p1, p2)
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

func (g *keyboardSearch) Apply(step string) bool {
	if child, ok := g.node.children[step]; ok {
		g.node = child
		return true
	}
	child := g.node.newChildKeyboardNode(step)
	g.node.children[step] = child
	g.node = child
	return true
}

func (g *keyboardSearch) Root() {
	g.node = g.root
}

func (g *keyboardSearch) Log() mcts.Log {
	return &keyboardLog{}
}

func (g *keyboardSearch) Expand() string {
	if g.node.depth >= 5 {
		return ""
	}
	p1, p2 := NewRandomValidPt(g.r), NewRandomValidPt(g.r)
	return fmt.Sprintf("%d %d %d %d", p1.X, p1.Y, p2.X, p2.Y)
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

	opts := mcts.Search{
		MinSelectBurnInDepth:     0,
		ExtraExpandBurnInSamples: 0,
		MaxExpandSamples:         1,
		RolloutsPerEpoch:         10,
		ExplorationParameter:     math.Pi,
	}
	res := opts.Search(s, done)

	fmt.Println(res)

	layout := NewRandomLayout(r)
	for leaf := res.PV; leaf != nil; leaf = leaf.PV {
		p1, p2 := parseStep(leaf.Step)
		layout.Swap(p1, p2)
	}

	fmt.Println(layout)
}
