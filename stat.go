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
	Step        S
	Log         Log
	Score       float64
	RawScore    float64
	NumRollouts float64
	Priority    float64
	Terminal    bool

	NumChildren      int
	NumExpandHits    float64
	NumExpandSamples float64
}

func makeStatEntry[S Step](n *heapordered.Tree[*node[S]]) StatEntry[S] {
	e, _ := n.Elem()
	return StatEntry[S]{
		Step:             e.Step,
		Log:              e.Log,
		RawScore:         e.Log.Score(),
		Score:            e.NormScore(),
		NumRollouts:      e.numRollouts,
		Priority:         e.priority,
		Terminal:         e.terminal,
		NumChildren:      n.Len(),
		NumExpandHits:    float64(e.hits),
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
func (r Search[S]) PV() Variation[S] { return nodeFuncV(r.root, r.Rand, mostNode) }

// AnyV returns a random variation with runs for this Search.
//
// AnyV is useful for statistical sampling of the Search tree.
func (r Search[S]) AnyV() Variation[S] { return nodeFuncV(r.root, r.Rand, anyNode) }

// FuncV creates a variation by calling f at every step.
func (r Search[S]) FuncV(f func([]StatEntry[S]) (int, bool)) Variation[S] {
	return nodeFuncV(r.root, r.Rand, func(nodes []*heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
		stat := make([]StatEntry[S], 0, len(nodes))
		for _, n := range nodes {
			stat = append(stat, makeStatEntry(n))
		}
		i, ok := f(stat)
		if !ok {
			return nil
		}
		return nodes[i]
	})
}

func nodeFuncV[S Step](root *heapordered.Tree[*node[S]], r *rand.Rand, f func([]*heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]]) Variation[S] {
	var res Variation[S]
	for n := root; n != nil; {
		res = append(res, makeStatEntry(n))
		e, _ := n.Elem()
		children := maps.Values(e.childSet)
		r.Shuffle(len(children), func(i, j int) { children[i], children[j] = children[j], children[i] })
		n = f(children)
	}
	return res
}

// Best returns the best Step for this Search root or nil.
func (r Search[S]) Best() *StatEntry[S] {
	if r.root == nil {
		return nil
	}
	e, _ := r.root.Elem()
	children := maps.Values(e.childSet)
	r.Rand.Shuffle(len(children), func(i, j int) { children[i], children[j] = children[j], children[i] })
	child := mostNode(children)
	stat := makeStatEntry(child)
	return &stat
}

// mostNode selects a node which maximizes runs.
//
// nodes should be shuffled prior to call.
func mostNode[S Step](nodes []*heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	var (
		maxNodes    []*heapordered.Tree[*node[S]]
		maxRollouts float64
	)
	for _, n := range nodes {
		e, _ := n.Elem()
		if e.numRollouts == maxRollouts {
			maxNodes = append(maxNodes, n)
		} else if e.numRollouts > maxRollouts {
			maxNodes = maxNodes[:0]
			maxNodes = append(maxNodes, n)
			maxRollouts = e.numRollouts
		}
	}
	if maxRollouts == 0 {
		// NormScore is -∞ when rollouts is 0.
		// Break the tie based on raw score.
		return bestRawScore(maxNodes)
	}
	return bestScore(maxNodes)
}

// bestScore picks the node with the best normed score.
//
// nodes should be shuffled prior to call.
func bestScore[S Step](nodes []*heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	var (
		maxNode  *heapordered.Tree[*node[S]]
		maxScore = math.Inf(-1)
	)
	for _, node := range nodes {
		e, _ := node.Elem()
		if score := e.NormScore(); maxNode == nil || score > maxScore {
			maxNode = node
			maxScore = score
		}
	}
	return maxNode
}

// bestRawScore picks the node with the best raw score.
//
// nodes should be shuffled prior to call.
func bestRawScore[S Step](nodes []*heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	var (
		maxNode  *heapordered.Tree[*node[S]]
		maxScore = math.Inf(-1)
	)
	for _, node := range nodes {
		e, _ := node.Elem()
		if score, _ := e.Score(); maxNode == nil || score > maxScore {
			maxNode = node
			maxScore = score
		}
	}
	return maxNode
}

// anyNode returns the first node with a nonzero number of runs or nil.
//
// Nodes should be shuffled prior to calling anyNode.
func anyNode[S Step](nodes []*heapordered.Tree[*node[S]]) *heapordered.Tree[*node[S]] {
	for _, n := range nodes {
		if e, _ := n.Elem(); e.numRollouts > 0 {
			return n
		}
	}
	return nil
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
