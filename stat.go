package mcts

import (
	"fmt"
	"strings"
)

type Stat struct {
	StatEntry
	PV *Stat
}

func (s Stat) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-2d %s", 0, s.StatEntry.String())
	for i, pv := 1, s.PV; pv != nil; i, pv = i+1, pv.PV {
		fmt.Fprintf(&sb, "%-2d %s", i, pv.StatEntry.String())
	}
	return sb.String()
}

type StatEntry struct {
	Step     Step
	EventLog EventLog
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

func (e *EventLog) prettyFormatExpandStats() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d / %d hits", e.NumExpandHits(), e.NumExpandSamples)
	if e.MaxSelectSamples > 0 {
		fmt.Fprintf(&sb, "; %d max", e.MaxSelectSamples)
	}
	return sb.String()
}

func (e StatEntry) String() string {
	return fmt.Sprintf("[%-4.3f] %-6s (%s; %s)\n",
		e.EventLog.Log.Score()/float64(e.EventLog.NumRollouts),
		e.Step,
		prettyFormatNumRollouts(e.EventLog.NumRollouts),
		e.EventLog.prettyFormatExpandStats(),
	)
}

func (n *EventLog) statEntry() StatEntry {
	score, _ := n.Score()
	return StatEntry{
		Step:     n.Step,
		EventLog: *n,
		Score:    score,
	}
}

func (n *EventLog) makeResult() Stat {
	root := Stat{}
	for stat := &root; ; {
		stat.StatEntry = n.statEntry()
		if n = n.bestChild(); n == nil {
			break
		}
		next := &Stat{}
		stat.PV = next
		stat = next
	}
	return root
}
