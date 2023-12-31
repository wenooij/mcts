package mcts

import (
	"fmt"
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
	Step     E
	EventLog EventLog[E]
	Score    float64
}

func prettyFormatNumRollouts(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d N", n)
	}
	if n < 1e6 {
		return fmt.Sprintf("%.2f kN", float64(n)/1e3)
	}
	return fmt.Sprintf("%.2f MN", float64(n)/1e6)
}

func (e *EventLog[E]) prettyFormatExpandStats() string {
	return fmt.Sprintf("%d children; %d samples", len(e.children), e.expandHeuristic.samples)
}

func (e StatEntry[E]) String() string {
	return fmt.Sprintf("[%-4.3f] %-6s (%s; %s)\n",
		e.EventLog.Log.Score()/float64(e.EventLog.NumRollouts),
		e.Step,
		prettyFormatNumRollouts(e.EventLog.NumRollouts),
		e.EventLog.prettyFormatExpandStats(),
	)
}

func (n *EventLog[E]) statEntry() StatEntry[E] {
	score, _ := n.Score()
	return StatEntry[E]{
		Step:     n.Step,
		EventLog: *n,
		Score:    score,
	}
}

func (n *EventLog[E]) makeResult(r *rand.Rand) Stat[E] {
	root := Stat[E]{}
	for stat := &root; ; {
		stat.StatEntry = n.statEntry()
		if n = n.bestChild(r); n == nil {
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
