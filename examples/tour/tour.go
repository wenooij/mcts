// Package tour implements a toy Travelling-Salesman solver.
// Use https://www.lancaster.ac.uk/fas/psych/software/TSP/TSP.html to validate the results.
package main

import (
	"flag"
	"fmt"
	"hash/maphash"
	"math"
	"math/rand"
	"slices"
	"strconv"
	"strings"
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
	expandN int
}

func newTourSearch(tourMap map[int]*tourPos, r *rand.Rand, tour string) *tourSearch {
	var root []int
	if tour != "" {
		for _, s := range strings.Split(tour, " ") {
			i, err := strconv.ParseInt(s, 10, 64) // May be 32 on some arch?
			if err != nil {
				panic(err)
			}
			root = append(root, int(i))
		}
	} else {
		root = rootTour(len(tourMap), r)
	}
	s := &tourSearch{
		m:    tourMap,
		r:    r,
		root: root,
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
	if g.node.depth >= 2*len(g.m) {
		return false
	}
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
	g.expandN = 0
}

func (g *tourSearch) Score() mcts.Score[float64] {
	score := mcts.Score[float64]{
		Counter:   -float64(g.node.depth),
		Objective: model.Minimize[float64],
	}
	if g.expandN == 0 {
		return score
	}
	score.Objective = nil
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
	score.Counter += distance
	return score
}

func (g *tourSearch) Expand(n int) []mcts.FrontierAction {
	if g.expandN = n; g.node.depth >= 2*len(g.m) {
		return nil
	}
	return g.actions
}

var seed = maphash.MakeSeed()

func (g *tourSearch) Hash() uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	for _, i := range g.node.tour {
		h.WriteByte(byte(i))
	}
	return h.Sum64()
}

func main() {
	tour := flag.String("tour", "", "When provided, seeds the initial tour with 1-indexed space delimited stops")
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	n := flag.Int("n", 10, "Number of tour stops")
	randomMap := flag.Bool("randomize_map", false, "Randomize the tour map")
	flag.Parse()

	r := rand.New(rand.NewSource(*seed))
	// Default tour for n=10.
	// Replace this with your own.
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
	if *n != 10 && *tour == "" && !*randomMap {
		fmt.Println("Setting -randomize_map=true when n != 10 and no initial -tour provided")
		*randomMap = true
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
	tourSearch := newTourSearch(tourMap, r, *tour)

	s := &mcts.Search[float64]{
		Rand:            r,
		Seed:            *seed,
		SearchInterface: model.MakeSearchInterface(tourSearch, mcts.CounterInterface[float64]{}),
		ExploreFactor:   600,
		NumEpisodes:     1000,
	}

	const epochs = 10
	start := time.Now()
	for i := 0; i < epochs; i++ {
		s.Search()
	}
	fmt.Println("Search took", time.Since(start), "using", len(s.Table), "table entries and", s.NumEpisodes*epochs, "iterations")

	// Reconstruct and print the best tour.
	// Results can be pasted into the tool.
	// https://www.lancaster.ac.uk/fas/psych/software/TSP/TSP.html.
	pv := searchops.FilterV(s.RootEntry,
		searchops.HasObjective[float64](),
		searchops.MaxScoreFilter[float64](),
		searchops.MaxDepthFilter[float64](2**n),
		searchops.FirstFilter[float64]())
	{
		tour := make([]int, len(tourSearch.root))
		copy(tour, tourSearch.root)
		for _, e := range pv {
			a := e.Action.(tourAction)
			tour[a.i], tour[a.j] = tour[a.j], tour[a.i]
		}
		fmt.Printf("[%f] ", pv[len(pv)-1].Score.Apply()/float64(pv[len(pv)-1].NumRollouts))
		for i, e := range tour {
			fmt.Printf("%d", e+1)
			if i+1 < len(tour) {
				fmt.Print(" ")
			}
		}
	}
	fmt.Printf(" (%f; %f; N=%d)\n", pv[0].NumRollouts, pv[len(pv)-1].NumRollouts, len(s.Table))
}
