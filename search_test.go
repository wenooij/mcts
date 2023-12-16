package mcts

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

type fakeTimer struct {
	wall int
	done chan struct{}
}

func newFakeTimer(t int) *fakeTimer {
	return &fakeTimer{wall: t, done: make(chan struct{})}
}

func (t *fakeTimer) Tick() {
	if t.wall--; t.wall <= 0 {
		t.wall = 0
		go func() { t.done <- struct{}{} }()
	}
}

type step int

func (s step) String() string {
	if s == 0 {
		return "#"
	}
	return strconv.FormatInt(int64(s), 10)
}

type node struct {
	depth    int
	value    float64
	children map[step]*node
}

func newRoot() *node {
	return &node{
		children: make(map[step]*node, 5),
	}
}

func (n *node) createChild(s step) *node {
	child := &node{
		depth:    n.depth + 1,
		value:    float64(n.value) + float64(s),
		children: make(map[step]*node, 5),
	}
	n.children[s] = child
	return child
}

type log struct {
	score float64
}

func (e *log) Merge(lg Log) {
	e.score += lg.(*log).score
}

func (e *log) Score() float64 {
	return e.score
}

type search struct {
	b    int
	d    int
	t    *fakeTimer
	r    *rand.Rand
	root *node
	node *node
}

func newSearch(t *testing.T, timer *fakeTimer, r *rand.Rand, b, d int) *search {
	t.Helper()
	root := newRoot()
	return &search{
		b:    b,
		d:    d,
		t:    timer,
		r:    r,
		root: root,
		node: root,
	}
}

func (s *search) Root() {
	s.node = s.root
	s.t.Tick()
}

func (s *search) Expand() step {
	if s.node.depth == s.d {
		return 0
	}
	return step(s.r.Intn(s.b) + 1)
}

func (s *search) Apply(step step) {
	child, ok := s.node.children[step]
	if !ok {
		child = s.node.createChild(step)
	}
	s.node = child
}

func (s *search) Log() Log {
	return new(log)
}

func (s *search) Rollout() Log {
	frontier := s.node
	defer func() { s.node = frontier }()
	log := new(log)
	for s.forward(log) {
	}
	return log
}

func (s *search) forward(log *log) bool {
	step := s.Expand()
	if step == 0 {
		log.score += s.node.value
		return false
	}
	s.Apply(step)
	return true
}

func TestSearchRecall(t *testing.T) {
	const seed = 1337
	for _, tc := range []struct {
		name          string
		inputBranches int
		inputDepth    int
		timeLimit     int
		wantRecall    float64
	}{{
		name:          "branching factor 2, depth 1, 1 epoch -> recall 100%",
		inputBranches: 2,
		inputDepth:    1,
		timeLimit:     1,
		wantRecall:    1,
	}, {
		name:          "branching factor 10, depth 10, 1000 epoch -> recall 100%",
		inputBranches: 2,
		inputDepth:    10,
		timeLimit:     1000,
		wantRecall:    1,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			timer := newFakeTimer(tc.timeLimit)
			s := newSearch(t, timer, r, tc.inputBranches, tc.inputDepth)
			c := Search[step]{Seed: 1337}
			res := c.Search(s, timer.done)
			bestLeaf := &res
			for ; bestLeaf.PV != nil; bestLeaf = bestLeaf.PV {
			}
			score := bestLeaf.Score / float64(bestLeaf.EventLog.NumRollouts)
			bestScore := float64(tc.inputDepth * tc.inputBranches)
			gotRecall := score / bestScore
			if gotRecall < 0 || gotRecall > 1 {
				t.Fatalf("TestSearchRecall(%q): recall %v is out of the expected range", tc.name, gotRecall)
			}
			if gotRecall < tc.wantRecall {
				t.Errorf("TestSearchRecall(%q): got recall %.2f%%, want recall %.2f%%", tc.name, 100*gotRecall, 100*tc.wantRecall)
			}
		})
	}

	done := make(chan struct{})
	timer := time.After(60 * time.Second)
	go func() {
		<-timer
		done <- struct{}{}
	}()
}
