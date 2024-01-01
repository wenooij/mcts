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

type testNode struct {
	depth    int
	value    float64
	children map[step]*testNode
}

func newRootForTest() *testNode {
	return &testNode{
		children: make(map[step]*testNode, 5),
	}
}

func (n *testNode) createChild(s step) *testNode {
	child := &testNode{
		depth:    n.depth + 1,
		value:    float64(n.value) + float64(s),
		children: make(map[step]*testNode, 5),
	}
	n.children[s] = child
	return child
}

type log struct {
	score float64
}

func (e *log) Merge(lg Log) Log {
	e.score += lg.(*log).score
	return e
}

func (e *log) Score() float64 {
	return e.score
}

type search struct {
	b    int
	d    int
	t    *fakeTimer
	r    *rand.Rand
	root *testNode
	node *testNode
}

func newSearch(t *testing.T, timer *fakeTimer, b, d int) *search {
	t.Helper()
	root := newRootForTest()
	r := rand.New(rand.NewSource(1337))
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

func (s *search) Rollout() (Log, int) {
	frontier := s.node
	defer func() { s.node = frontier }()
	log := new(log)
	for s.forward(log) {
	}
	return log, 1
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
	for _, tc := range []struct {
		name           string
		inputBranches  int
		inputDepth     int
		overrideBurnIn int
		timeLimit      int
		wantRecall     float64
	}{{
		name:          "{b:2, d:1, t:1}: 100%",
		inputBranches: 2,
		inputDepth:    1,
		timeLimit:     1,
		wantRecall:    1,
	}, {
		name:           "{b:100, d:1, t:1}: 100%",
		inputBranches:  100,
		inputDepth:     1,
		overrideBurnIn: 100,
		timeLimit:      1,
		wantRecall:     1,
	}, {
		name:          "{b:2, d:10, t:1000}: 85%",
		inputBranches: 2,
		inputDepth:    10,
		timeLimit:     1000,
		wantRecall:    .85,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			timer := newFakeTimer(tc.timeLimit)
			s := newSearch(t, timer, tc.inputBranches, tc.inputDepth)
			c := Search[step]{Seed: 13323427, ExpandBurnInSamples: tc.overrideBurnIn}
			res := c.Search(s, timer.done)
			bestLeaf := &res
			for ; bestLeaf.PV != nil; bestLeaf = bestLeaf.PV {
			}
			score := bestLeaf.Score / bestLeaf.LogNode.numRollouts
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
