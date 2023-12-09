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

func (n *logTree) statEntry() StatEntry {
	score, _ := n.eventLog.Score()
	return StatEntry{
		Step:     n.step,
		EventLog: n.eventLog,
		Score:    score,
	}
}

func (n *logTree) makeResult() Stat {
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
