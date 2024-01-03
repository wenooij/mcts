package main

import (
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/wenooij/mcts"
)

type tourStep int

func (s tourStep) String() string {
	if s == 0 {
		return "#"
	}
	return fmt.Sprintf("{%d}", int(s))
}

type tourDistanceLog float64

func (d tourDistanceLog) Score() float64             { return -float64(d) }
func (d tourDistanceLog) Merge(lg mcts.Log) mcts.Log { return d + lg.(tourDistanceLog) }

type tourPos struct {
	X int
	Y int
}

func makeTourPos(r *rand.Rand) tourPos {
	return tourPos{
		X: rand.Intn(100) - 50,
		Y: rand.Intn(100) - 50,
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func (p tourPos) DistanceTo(p2 tourPos) int {
	dx, dy := p2.X-p.X, p2.Y-p.Y
	return dx*dx + dy*dy
}

func makeTourMap(n int, r *rand.Rand) map[tourStep]tourPos {
	m := make(map[tourStep]tourPos, n)
	for i := tourStep(1); i <= tourStep(n); i++ {
		m[i] = makeTourPos(r)
	}
	return m
}

type tourNode struct {
	tourMap   map[tourStep]tourPos
	remaining []tourStep
	children  map[tourStep]*tourNode
	step      tourStep

	distance  tourDistanceLog
	expandPtr int
}

func newRootTourNode(n int, tourMap map[tourStep]tourPos, r *rand.Rand) *tourNode {
	return &tourNode{
		remaining: makeRemainingTourSteps(n, r),
		tourMap:   tourMap,
		children:  make(map[tourStep]*tourNode, n),
	}
}

func makeRemainingTourSteps(n int, r *rand.Rand) []tourStep {
	m := make([]tourStep, n)
	for i := 0; i < n; i++ {
		m[i] = tourStep(i + 1)
	}
	r.Shuffle(len(m), func(i, j int) { m[i], m[j] = m[j], m[i] })
	return m
}

func (n *tourNode) newChildTourNode(step tourStep, r *rand.Rand) *tourNode {
	child := &tourNode{
		tourMap:   n.tourMap,
		remaining: slices.Clone(n.remaining),
		children:  make(map[tourStep]*tourNode, len(n.remaining)-1),
		step:      step,
	}
	child.remaining = slices.DeleteFunc(child.remaining, func(i tourStep) bool { return i == step })
	r.Shuffle(len(child.remaining), func(i, j int) {
		child.remaining[i], child.remaining[j] = child.remaining[j], child.remaining[i]
	})
	child.distance = n.distance + tourDistanceLog(n.tourMap[n.step].DistanceTo(n.tourMap[step]))
	return child
}

type tourSearch struct {
	r    *rand.Rand
	root *tourNode
	node *tourNode
}

func newTourSearch(n int, tourMap map[tourStep]tourPos, r *rand.Rand) *tourSearch {
	return &tourSearch{
		r:    r,
		root: newRootTourNode(n, tourMap, r),
	}
}

func (g *tourSearch) Apply(step tourStep) {
	if child, ok := g.node.children[step]; ok {
		g.node = child
		return
	}
	child := g.node.newChildTourNode(step, g.r)
	g.node.children[step] = child
	g.node = child
}

func (g *tourSearch) Root() {
	g.node = g.root
}

func (g *tourSearch) Log() mcts.Log {
	return tourDistanceLog(0)
}

func (g *tourSearch) Expand() ([]tourStep, bool) {
	n := len(g.node.remaining)
	if n == 0 {
		return nil, true
	}
	step := g.node.remaining[g.node.expandPtr%n]
	g.node.expandPtr++
	return []tourStep{step}, false
}

func (g *tourSearch) Rollout() (mcts.Log, int) {
	for len(g.node.remaining) > 0 {
		idx := g.r.Intn(len(g.node.remaining))
		g.Apply(g.node.remaining[idx])
	}
	return tourDistanceLog(g.node.distance), 1
}

func main() {
	const n = 30
	const seed = 1337

	r := rand.New(rand.NewSource(seed))
	tourMap := makeTourMap(n, r)
	// map[tourStep]tourPos{
	// 	1: {-21, +45},
	// 	2: {+34, -38},
	// 	3: {+42, +37},
	// }
	s := newTourSearch(n, tourMap, r)

	done := make(chan struct{})
	go func() {
		<-time.After(10 * time.Second)
		done <- struct{}{}
	}()

	opts := mcts.Search[tourStep]{
		Seed:                     seed,
		ExpandBurnInSamples:      n,
		MaxSpeculativeExpansions: 1,
		ExplorationParameter:     2500,
		SearchInterface:          s,
		Done:                     done,
	}
	res := opts.Search()

	fmt.Println("map:")
	for i, p := range tourMap {
		fmt.Println(i, p)
	}

	for i := 1; i <= n; i++ {
		fmt.Println(i, opts.Score(tourStep(i)))
	}

	fmt.Println(res)
}
