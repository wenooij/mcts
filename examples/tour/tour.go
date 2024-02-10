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
	summ    *model.SummaryStats
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

func (g *tourSearch) Select(a mcts.Action) {
	g.node.apply(a.(tourAction))
}

func (g *tourSearch) Root() {
	if len(g.node.tour) == 0 {
		g.node.tour = make([]int, len(g.root))
	}
	copy(g.node.tour, g.root)
	g.node.depth = 0
	g.node.action = tourAction{}
}

func (g *tourSearch) Score() mcts.Score {
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
	if g.summ == nil {
		return model.Score(-distance) // Minimize distance.
	}
	return model.Score(g.summ.ZScore(-distance))
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
	s := newTourSearch(tourMap, r)

	opts := mcts.Search{
		Rand:            r,
		Seed:            *seed,
		SearchInterface: s,
		NumEpisodes:     100,
	}
	// summary := model.Summarize(&opts)
	// fmt.Println(summary.String())
	// s.summ = &summary
	epoch := 0
	const initTemp = 500
	opts.ExploreTemperature = initTemp
	for lastPrint := (time.Time{}); ; {
		if opts.Search(); time.Since(lastPrint) >= time.Second {
			// Reconstruct and print the best tour.
			// Results can be pasted into the tool.
			// https://www.lancaster.ac.uk/fas/psych/software/TSP/TSP.html.
			pv := opts.FilterV(mcts.MaxFilter(func(e *mcts.Node) float64 {
				if math.IsInf(e.Score(), 0) {
					return math.Inf(-1)
				}
				return e.Score()
			}), mcts.MaxRolloutsFilter(), mcts.AnyFilter(r))
			tour := make([]int, len(s.root))
			copy(tour, s.root)
			for _, e := range pv.TrimRoot() {
				a := e.Action().(tourAction)
				tour[a.i], tour[a.j] = tour[a.j], tour[a.i]
			}
			fmt.Printf("[%f] ", pv.Last().Score())
			for i, e := range tour {
				fmt.Printf("%d", e+1)
				if i+1 < len(tour) {
					fmt.Print(" ")
				}
			}
			fmt.Println()
			// fmt.Printf(" =(%f)\n", -pv.Last().Score*summary.Stddev-summary.Mean)

			lastPrint = time.Now()

			// Next epoch update temperature.
			epoch++
			if opts.ExploreTemperature <= 50 {
				fmt.Println("Reset temperature to start another pass")
				opts.Reset()
				opts.ExploreTemperature = initTemp
				opts.InsertV(pv.TrimRoot())
			} else if opts.ExploreTemperature <= 65 && opts.ExploreTemperature >= 1 {
				opts.ExploreTemperature *= .99
			} else if opts.ExploreTemperature <= 100 {
				opts.ExploreTemperature *= .98
			} else {
				opts.ExploreTemperature *= .85
			}
			fmt.Println("Temp is now: ", opts.ExploreTemperature)
		}
	}
}
