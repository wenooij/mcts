package main

import (
	"flag"
	"fmt"
	"hash/maphash"
	"math"
	"math/rand"
	"time"

	"github.com/wenooij/mcts"
	"github.com/wenooij/mcts/model"
	"github.com/wenooij/mcts/searchops"
)

type (
	shape []int
	bin   shape
	item  shape
)

func size(xs []int) int {
	var sum int
	for _, x := range xs {
		sum += x
	}
	return sum
}

func (i item) Size() int { return size([]int(i)) }
func (b bin) Size() int  { return size([]int(b)) }
func (b bin) CanPack(it item) bool {
	for i, cap := range b {
		if cap < it[i] {
			return false
		}
	}
	return true
}

type alloc struct {
	bin  int
	item int
}

func (a alloc) String() string { return fmt.Sprintf("{%d,%d}", a.bin, a.item) }

type binPacker struct {
	items            []item                // items.
	itemSizes        []int                 // item sizes.
	itemAllocated    []bool                // item allocated to a bin.
	numItemAllocated int                   // number of items allocated.
	binAllocated     []bool                // bin allocated with at least one item.
	bins             []bin                 // bins.
	rem              []bin                 // remaining bin space, can be negative.
	remBinSizes      []int                 // remaining bin sizes updated during Select.
	actions          []mcts.FrontierAction // a reusable Expand buffer.
	maxBinDim        int                   // maximim bin dimension value.
	numDims          int                   // num bin/item dimensions.
	r                *rand.Rand            // rng.
}

func (b *binPacker) Root() {
	b.numItemAllocated = 0      // Reset currItem.
	if b.itemAllocated == nil { // Allocate itemAllocated once.
		b.itemAllocated = make([]bool, len(b.items))
	}
	if b.binAllocated == nil { // Allocate binAllocated once.
		b.binAllocated = make([]bool, len(b.bins))
	}
	clear(b.itemAllocated)  // Clear  itemAllocated every call to Root.
	clear(b.binAllocated)   // Clear binAllocated every call to Root.
	if b.itemSizes == nil { // Allocate and initialize itemSizes once; it doesn't change.
		b.itemSizes = make([]int, len(b.items))
		for i := range b.itemSizes {
			b.itemSizes[i] = b.items[i].Size()
		}
	}
	if b.remBinSizes == nil { // Allocate remBinSizes once.
		b.remBinSizes = make([]int, len(b.bins))
	}
	if b.rem == nil { // Allocate rem once.
		b.rem = make([]bin, len(b.bins))
		for i := range b.rem {
			b.rem[i] = make(bin, b.numDims)
		}
	}
	// Initialize rem, remBinSizes on every call to Root.
	for i := range b.bins {
		copy(b.rem[i], b.bins[i])
		b.remBinSizes[i] = b.bins[i].Size()
	}
}

func (b *binPacker) Expand(n int) []mcts.FrontierAction {
	b.actions = b.actions[:0]
	if b.numItemAllocated >= len(b.items) {
		return nil // All items allocated.
	}
	binDimProduct := b.maxBinDim * b.numDims
	for i := 0; i < len(b.items); i++ {
		if b.itemAllocated[i] {
			continue // Skip i when item is allocated.
		}
		item := b.items[i]
		itemSize := b.itemSizes[i]
		for j := range b.bins {
			if !b.rem[j].CanPack(item) {
				continue // Skip unpackable items.
			}
			b.actions = append(b.actions, mcts.FrontierAction{
				Action: alloc{j, i},
				// Weight is assigned based on quality of fit.
				Weight: float64(binDimProduct - (b.remBinSizes[j] - itemSize)),
			})
		}
	}
	return b.actions
}

var seed = maphash.MakeSeed()

func (b *binPacker) Hash() uint64 {
	var h maphash.Hash
	h.SetSeed(seed)
	for i := range b.remBinSizes {
		size := b.remBinSizes[i]
		h.WriteByte(byte(size))
		h.WriteByte(byte(size >> 8))
		h.WriteByte(byte(size >> 16))
		h.WriteByte(byte(size >> 24))
	}
	return h.Sum64()
}

func (b *binPacker) Score() mcts.Score[int] {
	if b.numItemAllocated < len(b.items) {
		return mcts.Score[int]{Objective: model.Minimize[int]}
	}
	loss := b.loss()
	return mcts.Score[int]{Counter: loss, Objective: model.Minimize[int]}
}

func (b *binPacker) loss() int {
	var loss int
	for i := range b.remBinSizes {
		if b.binAllocated[i] {
			loss += b.remBinSizes[i]
		}
	}
	return loss
}

func (b *binPacker) Select(a mcts.Action) bool {
	aa := a.(alloc)
	b.itemAllocated[aa.item] = true
	b.binAllocated[aa.bin] = true
	bin := b.rem[aa.bin]
	item := b.items[aa.item]
	for i := range bin {
		bin[i] -= item[i]
	}
	b.remBinSizes[aa.bin] -= item.Size()
	b.numItemAllocated++
	return true
}

func (b *binPacker) Rollout() (int, float64) { return b.loss(), 1 }

func main() {
	exploreFactor := flag.Float64("c", math.Pi, "Explore factor for UCB")
	numBins := flag.Int("bins", 3, "Number of bin packing bins")
	numDims := flag.Int("dims", 3, "Number of bin pack dimensions")
	maxBinDim := flag.Int("max_bin_dim", 100, "Max bin dimension value")
	maxItemDim := flag.Int("max_item_dim", 10, "Max item dimension value")
	numItems := flag.Int("items", 9, "Number of bin pack items")
	seed := flag.Int64("seed", time.Now().UnixNano(), "Random seed")
	flag.Parse()

	fmt.Println("Using seed", *seed)

	if *maxBinDim >= 1<<32 {
		panic("-max_bin_dim is too big!")
	}

	r := rand.New(rand.NewSource(*seed))

	// Create items.
	items := make([]item, *numItems)
	for i := range items {
		item := make(item, *numDims)
		for j := range item {
			item[j] = r.Intn(*maxItemDim)
		}
		items[i] = item
		fmt.Println("item", i, item)
	}

	// Create bins.
	bins := make([]bin, *numBins)
	for i := range bins {
		bin := make(bin, *numDims)
		for j := range bin {
			bin[j] = r.Intn(*maxBinDim)
		}
		bins[i] = bin
		fmt.Println("bin", i, bin)
	}

	binPacker := &binPacker{
		items:     items,
		bins:      bins,
		maxBinDim: *maxBinDim,
		numDims:   *numDims,
		r:         r,
	}
	binPacker.Root()

	s := &mcts.Search[int]{
		SearchInterface: model.MakeSearchInterface(binPacker, mcts.CounterInterface[int]{}),
		ExploreFactor:   *exploreFactor,
		NumEpisodes:     100,
	}

	start := time.Now()
	lastPrint := (time.Time{})
	for epoch := 0; ; epoch++ {
		s.Search()
		if time.Since(lastPrint) > time.Second {
			fmt.Println("Search", time.Since(start), "using", len(s.Table), "table entries and", s.NumEpisodes*epoch, "iterations")
			pv := searchops.PV(s, searchops.MaxDepthFilter[int](*numItems))
			fmt.Println(pv)
			lastPrint = time.Now()
		}
	}
}
