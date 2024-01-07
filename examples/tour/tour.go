// Package tour implements a toy Travelling-Salesman solver.
// Use https://www.lancaster.ac.uk/fas/psych/software/TSP/TSP.html to validate the results.
package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
)

type tourStep struct {
	i int
	j int
}

func (s tourStep) String() string {
	if s.i == s.j {
		return "#"
	}
	return fmt.Sprintf("{%d,%d}", s.i, s.j)
}

type tourDistanceLog float64

func (d tourDistanceLog) Score() float64             { return -float64(d) }
func (d tourDistanceLog) Merge(lg mcts.Log) mcts.Log { return d + lg.(tourDistanceLog) }

type tourPos struct {
	X int
	Y int
}

func newTourPos(r *rand.Rand) *tourPos {
	return &tourPos{
		X: r.Intn(100) + 1,
		Y: r.Intn(100) + 1,
	}
}

func (p *tourPos) DistanceTo(p2 *tourPos) tourDistanceLog {
	dx, dy := p2.X-p.X, p2.Y-p.Y
	return tourDistanceLog(math.Sqrt(float64(dx*dx + dy*dy)))
}

func makeTourMap(n int, r *rand.Rand) map[int]*tourPos {
	m := make(map[int]*tourPos, n)
	for i := 0; i < n; i++ {
		m[i] = newTourPos(r)
	}
	r.Shuffle(len(m), func(i, j int) { m[i], m[j] = m[j], m[i] })
	return m
}

type tourNode struct {
	tour  []int
	step  tourStep
	depth int
}

func rootTour(n int) []int {
	tour := make([]int, n)
	for i := 0; i < n; i++ {
		tour[i] = i
	}
	return tour
}

func (n *tourNode) applyChildTourNode(tourMap map[int]*tourPos, step tourStep) {
	n.tour[step.i], n.tour[step.j] = n.tour[step.j], n.tour[step.i]
	n.step = step
	n.depth++
}

type tourSearch struct {
	m    map[int]*tourPos
	r    *rand.Rand
	root []int
	node *tourNode
}

func newTourSearch(tourMap map[int]*tourPos, r *rand.Rand) *tourSearch {
	s := &tourSearch{
		m:    tourMap,
		r:    r,
		root: rootTour(len(tourMap)),
		node: new(tourNode),
	}
	s.Root()
	return s
}

func (g *tourSearch) Apply(step tourStep) {
	g.node.applyChildTourNode(g.m, step)
}

func (g *tourSearch) Root() {
	if len(g.node.tour) == 0 {
		g.node.tour = make([]int, len(g.root))
	}
	copy(g.node.tour, g.root)
	g.node.depth = 0
	g.node.step = tourStep{}
}

func (g *tourSearch) Log() mcts.Log {
	return tourDistanceLog(0)
}

func (g *tourSearch) Expand() ([]tourStep, bool) {
	if g.node.depth >= len(g.node.tour)/2+1 {
		return nil, true
	}
	steps := make([]tourStep, 0, len(g.node.tour))
	i := g.node.tour[g.node.depth]
	for j := 0; j < len(g.node.tour); j++ {
		steps = append(steps, tourStep{i, j})
	}
	return steps, false
}

func (g *tourSearch) Rollout() (mcts.Log, int) {
	for {
		steps, terminal := g.Expand()
		if terminal {
			break
		}
		g.Apply(steps[rand.Intn(len(steps))])
	}
	// Calculate tour distance.
	distance := tourDistanceLog(0)
	first := g.m[g.node.tour[0]]
	last := first
	for _, e := range g.node.tour[1:] {
		curr := g.m[e]
		distance += last.DistanceTo(curr)
		last = curr
	}
	distance += last.DistanceTo(first)
	return distance, 1
}

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	randomMap := flag.Bool("randomize_map", false, "Randomize the tour map")
	flag.Parse()

	r := rand.New(rand.NewSource(*seed))
	pos := []*tourPos{
		{54, 66},
		{34, 29},
		{30, 31},
		{45, 54},
		{72, 47},
		{30, 7},
		{46, 62},
		{36, 84},
		{13, 81},
		{68, 69},
	}
	n := len(pos)
	tourMap := make(map[int]*tourPos, n)
	for i, p := range pos {
		tourMap[i] = p
	}
	if *randomMap {
		tourMap = makeTourMap(10, r)
	}
	for i := 0; i < n; i++ {
		p := tourMap[i]
		fmt.Printf("%d,%d\n", p.X, p.Y)
	}
	fmt.Println("---")
	s := newTourSearch(tourMap, r)

	done := make(chan struct{})
	go func() {
		<-time.After(60 * time.Second)
		done <- struct{}{}
	}()

	opts := mcts.Search[tourStep]{
		Rand:                     r,
		Seed:                     *seed,
		ExpandBurnInSamples:      1,
		MaxSpeculativeExpansions: 1,
		// InitialNodePriority:      100,
		SearchInterface: s,
		Done:            done,
	}
	model.FitParams(&opts)
	fmt.Printf("Using c=%.4f\n---\n", opts.ExplorationParameter)
	opts.Search()

	pv := opts.PV()
	fmt.Println(pv)
	fmt.Println("---")

	// Reconstruct and print the best tour.
	// Results can be pasted into the tool.
	// https://www.lancaster.ac.uk/fas/psych/software/TSP/TSP.html.
	tour := make([]int, len(s.root))
	copy(tour, s.root)
	for _, e := range pv {
		tour[e.Step.i], tour[e.Step.j] = tour[e.Step.j], tour[e.Step.i]
	}
	for i, e := range tour {
		fmt.Printf("%d", e+1)
		if i+1 < len(tour) {
			fmt.Print(" ")
		}
	}
	fmt.Println()
	fmt.Println("---")
	fmt.Println(-pv.Leaf().Score)
}
