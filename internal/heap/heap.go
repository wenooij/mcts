package heap

import "github.com/wenooij/mcts"

func Swap[T mcts.Counter](h []*mcts.Edge[T], i, j int) { h[i], h[j] = h[j], h[i] }

func Down[T mcts.Counter](h []*mcts.Edge[T], i0 int, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1                                                      // Left child.
		if j2 := j1 + 1; j2 < n && h[j2].Priority < h[j1].Priority { // Less(j2, j1)
			j = j2 // = 2*i + 2  // right child
		}
		if h[j].Priority >= h[i].Priority { // !Less(j, i)
			break
		}
		Swap(h, i, j)
		i = j
	}
	return i > i0
}
