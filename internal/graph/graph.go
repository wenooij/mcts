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

	// InverseTable is only used for the default Hash implementation.
	//
	// NOTE: Key of *EdgeList prevents EdgeLists from changing freely.
	InverseTable map[*mcts.EdgeList[T]]uint64
	m            maphash.Hash
}

func InternalInterface[T mcts.Counter]() mcts.InternalInterface[T] {
	g := &graphInterface[T]{}
	return mcts.InternalInterface[T]{
		Init:        g.init,
		Reset:       g.reset,
		Backprop:    backprop[T],
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
			if len(s.ForwardPath) == 0 {
				return g.m.Sum64()
			}
			binary.BigEndian.PutUint64(b[:], g.InverseTable[s.ForwardPath[len(s.ForwardPath)-1].Src])
			g.m.Write(b[:])
			g.m.WriteString(s.ForwardPath[len(s.ForwardPath)-1].Node.Action.String())
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
