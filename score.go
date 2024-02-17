package mcts

type Counter interface{}

type Score[T Counter] struct {
	Counter   T
	Objective func(T) float64
}

func (s Score[T]) Apply() float64 { return s.Objective(s.Counter) }
