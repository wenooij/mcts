// Package graph provides an internal interface for the graph model topology.
package graph

import (
	"encoding/binary"
	"hash/maphash"

	"github.com/wenooij/mcts"
)

type graphInterface[T mcts.Counter] struct {
	// Table is the collection of Hashed Nodes and children.
	Table map[uint64]*mcts.EdgeList[T]

	ForwardPath []*mcts.Edge[T]

	// InverseTable is only used for the default Hash implementation.
	//
	// NOTE: Key of *EdgeList prevents EdgeLists from changing freely.
	InverseTable map[*mcts.EdgeList[T]]uint64
	m            maphash.Hash
}

// SearchInterface wraps the search interface using the internal graph topoology.
func SearchInterface[T mcts.Counter](s mcts.SearchInterface[T]) mcts.SearchInterface[T] {
	g := &graphInterface[T]{}
	_ = g
	s.InternalInterface = g.InternalInterface()
	return s
}

func (g *graphInterface[T]) InternalInterface() mcts.InternalInterface[T] {
	return mcts.InternalInterface[T]{
		Init:        g.init,
		Reset:       g.reset,
		Root:        g.Root,
		Backprop:    g.backprop,
		Rollout:     rollout[T],
		Expand:      g.expand,
		SelectChild: g.selectChild,
		MakeNode:    makeNode[T],
	}
}

func (g *graphInterface[T]) reset(s *mcts.Search[T]) {
	g.Table = make(map[uint64]*mcts.EdgeList[T], 64)
	if g.InverseTable != nil {
		g.InverseTable = nil
		s.Hash = nil
	}
}

func (g *graphInterface[T]) init(s *mcts.Search[T]) {
	if g.Table == nil {
		g.Table = make(map[uint64]*mcts.EdgeList[T], 64)
	}
	if s.Hash == nil {
		g.InverseTable = make(map[*mcts.EdgeList[T]]uint64, 64)
		g.m.SetSeed(maphash.MakeSeed())
		var b [8]byte
		// Provide a default hash implementation which hashes the last state and the next move.
		s.Hash = func() uint64 {
			g.m.Reset()
			if len(g.ForwardPath) == 0 {
				return g.m.Sum64()
			}
			binary.BigEndian.PutUint64(b[:], g.InverseTable[g.ForwardPath[len(g.ForwardPath)-1].Src])
			g.m.Write(b[:])
			g.m.WriteString(g.ForwardPath[len(g.ForwardPath)-1].Node.Action.String())
			return g.m.Sum64()
		}
	}
	// Find the root hash node.
	if s.RootEntry == nil {
		s.Root()
		h := s.Hash()
		e, ok := g.Table[h]
		if !ok {
			// Initialize root.
			e = &mcts.EdgeList[T]{}
			g.Table[h] = e
			if g.InverseTable != nil {
				g.InverseTable[e] = h
			}
		}
		s.RootEntry = e
	}
}

func (g *graphInterface[T]) Root() {
	g.ForwardPath = g.ForwardPath[:0]
}
