package mcts

type rolloutRunner[S Step] struct{}

func (r *rolloutRunner[S]) Rollout(si SearchInterface[S]) (log Log, numRollouts int) {
	return si.Rollout()
}
