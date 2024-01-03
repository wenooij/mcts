package mcts

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

type Stat[E Step] struct {
	StatEntry[E]
	PV *Stat[E]
}

func (s Stat[E]) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-2d %s", 0, s.StatEntry.String())
	for i, pv := 1, s.PV; pv != nil; i, pv = i+1, pv.PV {
		fmt.Fprintf(&sb, "%-2d %s", i, pv.StatEntry.String())
	}
	return sb.String()
}

type StatEntry[E Step] struct {
	Step    E
	LogNode topo[E]
	Score   float64
}

func prettyFormatNumRollouts(n float64) string {
	if n < 1000 {
		return fmt.Sprintf("%.0f N", n)
	}
	if n < 1e6 {
		return fmt.Sprintf("%.2f kN", n/1e3)
	}
	return fmt.Sprintf("%.2f MN", n/1e6)
}

func (e *topo[E]) prettyFormatExpandStats() string {
	return fmt.Sprintf("%d children; %d samples", len(e.children), e.Samples())
}

func (e StatEntry[E]) String() string {
	return fmt.Sprintf("[%-4.3f] %-6s (%s; %s)\n",
		e.LogNode.Log.Score()/float64(e.LogNode.numRollouts),
		e.Step,
		prettyFormatNumRollouts(e.LogNode.numRollouts),
		e.LogNode.prettyFormatExpandStats(),
	)
}

func (n *topo[E]) statEntry() StatEntry[E] {
	score, _ := n.Score()
	return StatEntry[E]{
		Step:    n.Step,
		LogNode: *n,
		Score:   score,
	}
}

func (n *topo[E]) makeResult(r *rand.Rand) Stat[E] {
	root := Stat[E]{}
	for stat := &root; ; {
		stat.StatEntry = n.statEntry()
		if n = n.mostChild(r); n == nil {
			break
		}
		next := &Stat[E]{}
		stat.PV = next
		stat = next
	}
	return root
}

func (r Search[E]) Score(pv ...E) []float64 {
	node := r.root
	res := make([]float64, 0, 1+len(pv))
	for i := 0; ; i++ {
		e := node.statEntry()
		res = append(res, e.Score)
		if i >= len(pv) {
			break
		}
		s := pv[i]
		child, ok := node.childSet[s]
		if !ok {
			break
		}
		node = child
	}
	return res
}

func (log *topo[S]) mostChild(r *rand.Rand) *topo[S] {
	if log == nil || len(log.children) == 0 {
		return nil
	}
	// Select an existing child to maximize runs.
	var (
		maxChildren []*topo[S]
		maxRuns     = -1.0
	)
	for _, e := range log.children {
		if maxRuns < e.numRollouts {
			maxChildren = append(maxChildren[:0], e)
			maxRuns = e.numRollouts
		} else if maxRuns == e.numRollouts {
			maxChildren = append(maxChildren, e)
		}
	}
	return bestChild(r, maxChildren)
}

func bestChild[S Step](r *rand.Rand, children []*topo[S]) *topo[S] {
	if len(children) == 0 {
		return nil
	}
	if len(children) == 1 {
		return children[0]
	}
	var (
		maxChild *topo[S]
		maxScore = math.Inf(-1)
	)
	// Don't rely on the default map ordering.
	r.Shuffle(len(children), func(i, j int) { children[i], children[j] = children[j], children[i] })
	for _, e := range children {
		score, _ := e.Score()
		score /= float64(e.numRollouts)
		if maxScore < score {
			maxChild = e
			maxScore = score
		}
	}
	return maxChild
}
