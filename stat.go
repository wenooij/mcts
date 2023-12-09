package mcts

type Stat struct {
	StatEntry
	PV *Stat
}

type StatEntry struct {
	Step     Step
	EventLog EventLog
	Score    float64
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
