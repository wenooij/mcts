package mcts

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/wenooij/heapordered"
	"golang.org/x/exp/maps"
)

type StatEntry[S Step] struct {
	Step             S
	Score            float64
	RawScore         float64
	NumRollouts      float64
	NumChildren      int
	NumExpandSamples float64
}

func makeStatEntry[S Step](n *heapordered.Tree[*node[S]]) StatEntry[S] {
	e, _ := n.Elem()
	return StatEntry[S]{
		Step:             e.Step,
		RawScore:         e.Log.Score(),
		Score:            e.NormScore(),
		NumRollouts:      e.numRollouts,
		NumChildren:      n.Len(),
		NumExpandSamples: float64(e.Samples()),
	}
}

func (e StatEntry[S]) String() string {
	var sb strings.Builder
	e.appendString(&sb)
	return sb.String()
}

func (e StatEntry[S]) appendString(sb *strings.Builder) {
	fmt.Fprintf(sb, "[%-4.3f] %-6s (", e.Score, e.Step)
	// Format NumRollouts.
	switch n := e.NumRollouts; {
	case n < 1000:
		fmt.Fprintf(sb, "%.0f N; ", n)
	case n < 1e6:
		fmt.Fprintf(sb, "%.2f kN; ", n/1e3)
	default:
		fmt.Fprintf(sb, "%.2f MN; ", n/1e6)
	}
	// Format expand stats.
	fmt.Fprintf(sb, "%d children; %d samples)", e.NumChildren, int(e.NumExpandSamples))
}

// PV returns the principal variation for this Search.
//
// This is the line that the Search has searched deepest
// and is usually the best one.
//
// Use Stat to test arbitrary sequences.
func (r Search[S]) PV() Variation[S] {
	return makePV(r.root, r.Rand)
}

// AnyV returns stats for a random variation for this Search.
//
// AnyV is useful for statistical sampling of the Search tree.
func (r Search[S]) AnyV() Variation[S] {
	return makeAnyV(r.root, r.Rand)
}

// Best returns the best Step for this Search root or nil.
func (r Search[S]) Best() *StatEntry[S] {
	if r.root == nil {
		return nil
	}
	child := mostChild(r.root, r.Rand)
	e := makeStatEntry(child)
	return &e
}

func makePV[S Step](root *heapordered.Tree[*node[S]], r *rand.Rand) Variation[S] {
	var pv Variation[S]
	pv = append(pv, makeStatEntry(root))
	for root != nil {
		child := mostChild(root, r)
		if child == nil {
			break
		}
		pv = append(pv, makeStatEntry(child))
		root = child
	}
	return pv
}

func mostChild[S Step](n *heapordered.Tree[*node[S]], r *rand.Rand) *heapordered.Tree[*node[S]] {
	if n == nil || n.Len() == 0 {
		return nil
	}
	// Select an existing child to maximize runs.
	var (
		maxChildren []*heapordered.Tree[*node[S]]
		maxRollouts float64
	)
	e, _ := n.Elem()
	children := maps.Values(e.childSet)
	r.Shuffle(len(children), func(i, j int) { children[i], children[j] = children[j], children[i] })
	for _, child := range children {
		e, _ := child.Elem()
		if e.numRollouts == maxRollouts {
			maxChildren = append(maxChildren, child)
		} else if e.numRollouts > maxRollouts {
			maxChildren = maxChildren[:0]
			maxChildren = append(maxChildren, child)
			maxRollouts = e.numRollouts
		}
	}
	if maxRollouts == 0 && len(maxChildren) > 0 {
		// Random choice based on no information.
		// Pick the first one.
		return maxChildren[0]
	}
	return bestNode(r, maxChildren)
}

func bestNode[S Step](r *rand.Rand, nodes []*heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	switch len(nodes) {
	case 0:
		return nil
	case 1:
		return nodes[0]
	}
	var (
		maxChild *heapordered.Tree[*node[S]]
		maxScore = math.Inf(-1)
	)
	// Don't rely on the default map ordering.
	r.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })
	for _, node := range nodes {
		e, _ := node.Elem()
		score, ok := e.Score()
		if ok && e.numRollouts != 0 {
			score /= float64(e.numRollouts)
		} else {
			score = math.Inf(-1)
		}
		if maxChild == nil || score > maxScore {
			maxChild = node
			maxScore = score
		}
	}
	return maxChild
}

func makeAnyV[S Step](root *heapordered.Tree[*node[S]], r *rand.Rand) Variation[S] {
	var av Variation[S]
	av = append(av, makeStatEntry(root))
	for root != nil {
		child := anyChild(root, r)
		if child == nil {
			break
		}
		av = append(av, makeStatEntry(child))
		root = child
	}
	return av
}

func anyChild[S Step](root *heapordered.Tree[*node[S]], r *rand.Rand) *heapordered.Tree[*node[S]] {
	switch root.Len() {
	case 0:
		return nil
	case 1:
		return root.Min()
	}
	e, _ := root.Elem()
	keys := maps.Keys(e.childSet)
	return e.childSet[keys[r.Intn(len(keys))]]
}

// Variation is a sequence of Steps with Search statistics.
//
// The first element in the Variation is the root mode and will have the zero value for the Step.
type Variation[S Step] []StatEntry[S]

func (v Variation[S]) Leaf() *StatEntry[S] {
	if len(v) == 0 {
		return nil
	}
	leaf := v[len(v)-1]
	return &leaf
}

func (v Variation[S]) String() string {
	var sb strings.Builder
	if len(v) == 0 {
		return "\n"
	}
	for i := 0; i < len(v); i++ {
		e := v[i]
		fmt.Fprintf(&sb, "%-2d ", i)
		e.appendString(&sb)
		fmt.Fprintln(&sb)
	}
	return sb.String()
}

// Stat returns a sequence of Search stats for the given variation according to this Search.
//
// Stat will return all a slice of StatEntries equal to one plus the length of the input vs.
// If Search did not encounter those steps yet, the NumRollouts value will be 0.
func (r Search[S]) Stat(vs ...S) Variation[S] {
	n := r.root
	res := make(Variation[S], 0, 1+len(vs))
	res = append(res, makeStatEntry(n))
	for i, s := range vs {
		child := getChild(n, s)
		if child == nil {
			// No existing child.
			// Add dummy stat entries and break.
			for _, s := range vs[i:] {
				res = append(res, StatEntry[S]{Step: s})
			}
			break
		}
		// Add the StatEntry and continue down the line.
		n = child
		res = append(res, makeStatEntry(n))
	}
	return res
}
