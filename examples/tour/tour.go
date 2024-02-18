// Package tour implements a toy Travelling-Salesman solver.
// Use https://www.lancaster.ac.uk/fas/psych/software/TSP/TSP.html to validate the results.
package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/searchops"
)

type tourAction struct {
	i int
	j int
}

func (s tourAction) String() string {
	if s.i == s.j {
		return "#"
	}
	return fmt.Sprintf("{%d,%d}", s.i, s.j)
}

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

func (p *tourPos) DistanceTo(p2 *tourPos) float64 {
	dx, dy := p2.X-p.X, p2.Y-p.Y
	return math.Sqrt(float64(dx*dx + dy*dy))
}

func makeTourMap(n int, r *rand.Rand) map[int]*tourPos {
	m := make(map[int]*tourPos, n)
	for i := 0; i < n; i++ {
		m[i] = newTourPos(r)
	}
	return m
}

type tourNode struct {
	tour   []int
	action tourAction
	depth  int
}

func rootTour(n int, r *rand.Rand) []int {
	tour := make([]int, n)
	for i := 0; i < n; i++ {
		tour[i] = i
	}
	r.Shuffle(len(tour), func(i, j int) { tour[i], tour[j] = tour[j], tour[i] })
	return tour
}

func (n *tourNode) apply(a tourAction) {
	n.tour[a.i], n.tour[a.j] = n.tour[a.j], n.tour[a.i]
	n.action = a
	n.depth++
}

type tourSearch struct {
	m       map[int]*tourPos
	r       *rand.Rand
	root    []int
	actions []mcts.FrontierAction
	node    *tourNode
}

func newTourSearch(tourMap map[int]*tourPos, r *rand.Rand) *tourSearch {
	s := &tourSearch{
		m:    tourMap,
		r:    r,
		root: rootTour(len(tourMap), r),
		node: new(tourNode),
	}
	s.actions = slices.Grow(s.actions, len(tourMap))
	for i := 0; i < len(s.root); i++ {
		for j := 0; j < len(s.root); j++ {
			if i == j {
				continue
			}
			s.actions = append(s.actions, mcts.FrontierAction{Action: tourAction{i, j}})
		}
	}
	s.Root()
	return s
}

func (g *tourSearch) Select(a mcts.Action) bool {
	g.node.apply(a.(tourAction))
	return true
}

func (g *tourSearch) Root() {
	if len(g.node.tour) == 0 {
		g.node.tour = make([]int, len(g.root))
	}
	copy(g.node.tour, g.root)
	g.node.depth = 0
	g.node.action = tourAction{}
}

func (g *tourSearch) Score() mcts.Score[float64] {
	// Calculate tour distance.
	distance := float64(0)
	first := g.m[g.node.tour[0]]
	last := first
	for _, e := range g.node.tour[1:] {
		curr := g.m[e]
		distance += last.DistanceTo(curr)
		last = curr
	}
	distance += last.DistanceTo(first)
	return mcts.Score[float64]{
		Counter:   distance,
		Objective: model.Minimize[float64],
	}
}

func (g *tourSearch) Expand(int) []mcts.FrontierAction {
	if g.node.depth >= 2*len(g.m) {
		return nil
	}
	return g.actions
}

func main() {
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	n := flag.Int("n", 10, "Number of tour stops")
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
	tourMap := make(map[int]*tourPos, *n)
	for i, p := range pos {
		tourMap[i] = p
	}
	if *randomMap {
		tourMap = makeTourMap(*n, r)
	}
	for i := 0; i < *n; i++ {
		p := tourMap[i]
		fmt.Printf("%d,%d\n", p.X, p.Y)
	}
	fmt.Println("---")
	tourSearch := newTourSearch(tourMap, r)

	s := mcts.Search[float64]{
		Rand:            r,
		Seed:            *seed,
		SearchInterface: tourSearch,
		NumEpisodes:     100,
	}
	for lastPrint := (time.Time{}); ; {
		if s.Search(); time.Since(lastPrint) >= time.Second {
			// Reconstruct and print the best tour.
			// Results can be pasted into the tool.
			// https://www.lancaster.ac.uk/fas/psych/software/TSP/TSP.html.
			pv := searchops.PV(s.Tree)
			tour := make([]int, len(tourSearch.root))
			copy(tour, tourSearch.root)
			for _, e := range pv.TrimRoot() {
				a := e.Action.(tourAction)
				tour[a.i], tour[a.j] = tour[a.j], tour[a.i]
			}
			fmt.Printf("[%f] ", pv.Last().Score.Apply()/float64(pv.Last().NumRollouts))
			for i, e := range tour {
				fmt.Printf("%d", e+1)
				if i+1 < len(tour) {
					fmt.Print(" ")
				}
			}
			fmt.Printf(" (%f)\n", pv[0].NumRollouts)

			lastPrint = time.Now()
		}
	}
}
