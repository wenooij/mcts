package mcts

type expandBurnIn[S Step] struct {
	*expander[S]

	expandBurnInSamples int
	burnedIn            bool
}

func (e *expandBurnIn[S]) Init(ex *expander[S], expandBurnInSamples int) {
	e.expander = ex
	e.expandBurnInSamples = expandBurnInSamples
}

func (e *expandBurnIn[S]) TryBurnIn() bool {
	if e.burnedIn {
		return false
	}
	for i := 0; i < e.expandBurnInSamples; i++ {
		e.expander.Expand()
	}
	e.burnedIn = true
	return true
}
