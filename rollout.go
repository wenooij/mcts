package mcts

type rolloutRunner[S Step] struct {
	SearchInterface[S]
	frontier *node[S]

	rolloutsPerEpoch int
}

func (r *rolloutRunner[S]) Init(si SearchInterface[S], frontier *node[S], rolloutsPerEpoch int) {
	r.SearchInterface = si
	r.frontier = frontier
	r.rolloutsPerEpoch = rolloutsPerEpoch
}

func (r *rolloutRunner[S]) Rollout() (log Log, numRollouts int) {
	log = r.SearchInterface.Log()
	for i := 0; i < r.rolloutsPerEpoch; i++ {
		if res := r.SearchInterface.Rollout(); res != nil {
			log = log.Merge(res)
			numRollouts++
		}
	}
	return log, numRollouts
}
