package graph

import "github.com/wenooij/mcts"

// selectChild selects the highest priority child from the min heap.
func selectChild[T mcts.Counter](s mcts.SearchInterface[T], table map[uint64]*mcts.TableEntry[T], hashTable map[*mcts.TableEntry[T]]uint64, trajectory *[]*mcts.Edge[T], n *mcts.TableEntry[T]) (child *mcts.Edge[T], expand bool) {
	if len(*n) == 0 {
		return nil, true
	}
	child = (*n)[0]
	if !s.Select(child.Action) {
		// Select may return false if this node is no longer legal
		// Possibly due to the outcome of chance node higher up the tree.
		// In SearchHash, Select may return false after a cycle is detected
		// or after a maximum depth is reached.
		//
		// In either case return child = nil, expand = false, then
		// backprop the score from n.
		return nil, false
	}
	*trajectory = append(*trajectory, child)
	if child.Dst == nil {
		// Insert initial node.
		// We couldn't do this in Expand because Hash
		// expects to be called only after Select.
		h := s.Hash()
		// Dst will already be in Table if dst is a transposition.
		dst, ok := table[h]
		if !ok {
			dst = &mcts.TableEntry[T]{}
			table[h] = dst
			if hashTable != nil {
				hashTable[dst] = h
			}
		}
		child.Dst = dst
	}
	initializeScore(s, child)
	return child, false
}

// initializeScore is called when selecting a node for the first time.
//
// precondition: n must be the current node (s.Select(n.Action) has been called, or we are at the root).
func initializeScore[T mcts.Counter](s mcts.SearchInterface[T], e *mcts.Edge[T]) {
	if e.Score.Objective == nil {
		// E will be heapified on the first call to backprop.
		e.Score = s.Score()
	}
}
