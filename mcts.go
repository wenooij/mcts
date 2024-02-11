// Package mcts provides an implementation of general multi-agent Monte-Carlo tree search (MCTS).
package mcts

type Score struct {
	Counters []float64
	ObjectiveFunc
}

func (s Score) Add(s2 Score) {
	for i, v := range s2.Counters {
		s.Counters[i] += v
	}
}

func (s Score) Apply() float64 { return s.ObjectiveFunc(s.Counters) }

// Action represents an edge in the a game tree.
//
// String should return a standard representation of the Action.
type Action interface {
	String() string
}

// FrontierAction wraps an Action returned from Expand with extra parameters to apply to its subtree.
//
// Weight is an optional predictor used to bias the MAB policy as described in
// <Rosin, Christopher D. "Multi-armed bandits with episode context."
// Annals of Mathematics and Artificial Intelligence 61.3 (2011)>
// Ideally, the predictor weight for X should be set to E[Mean(X_Score)].
// The weight heuristic is usually tuned in an offline process.
// Weight 0 will be smoothed to 1.
type FrontierAction struct {
	Action Action
	Weight float64
}

// SearchInterface is the minimal interface to MCTS tree state.
//
// SearchInterface provides a simple flat interface:
//
//   - Select moves the state of the game forward by applying the Action.
//   - Expand reveals available actions at the current state.
//   - Root reset the game tree to the state at the beginning of search.
//
// The game tree is managed by this package, so it is not required to implement
// one yourself, so long as Select and Root are relatively inexpensive.
// Expand will not be called again for nodes in the game tree.
//
// MCTS proceeds from the root, down the game tree, selecting the best action
// at each level, until it reaches a frontier leaf node where a random rollout
// is started. The implementor has two choices for the rollout:
//
//  1. Implement ScoreInterface and rely on the defeault rollout strategy.
//  2. Implement RolloutInterface and use a custom strategy.
//
// Which one to use will depend on one's needs:
//
// RolloutInterface allows for a custom implementation of rollouts.
// This may be useful for those who wish to apply custom heuristics for sampling actions.
// Additionally, this may be a good choice when Expand is expensive.
// Lastly, RolloutInterface supports multiple simulations per epoch
// which may provide further flexibility. For instance, running extra
// simulations for promising positions.
//
// Otherwise, the default rollout strategy can be used. This uses the base methods from
// SearchInterface, calling Select and Expand repeatedly until a terminal state has been
// reached (Expand returns an empty slice of actions).
// The resulting score from ScoreInterface.Score is returned and the rollout completes.
// If the implementation does not satisfy ScoreInterface, a random normally distributed
// dummy score is returned.
// The default rollout strategy may be good enough for one's needs and one done not have
// to bother implementing a custom strategy.
// Alternative implementations of Expand which apply heuristics or generates only a single
// actions must be carefully coded. For instance, failure to return available actions randomly
// will introduce biases to search. It has also been reported that optimal play in simulation
// may actually hinder the explorative performance of MCTS.
// Note that Expand can be made less expensive by reusing the same slice.
// The slice will not be retained by the implementation in this package.
type SearchInterface interface {
	// Root resets the current search to root.
	//
	// Root is called multiple times in Search before the selection phase
	// and after Search completes.
	Root()

	// Select applies the given Action.
	//
	// Select is called multiple times during the selection phase.
	// Select will also be called during rollout if Search does not implement RolloutInterface.
	Select(Action)

	// Expand returns at most n available actions.
	// When n <= 0, all available actions are returned.
	//
	// Expand is called after the selection phase with n <= 0 to expand the frontier of a leaf node.
	// If Expand returns no actions, the current state is marked as a terminal.
	//
	// Expand will be called during rollout with n = 1 if Search does not implement RolloutInterface.
	// Expand must always eventually return a terminal if using the default rollout strategy.
	Expand(n int) []FrontierAction

	// Score is a record for scorekeeping in search.
	//
	// Score will be called on each expanded node and on each terminal state reached.
	//
	// Score represents the objective score for the terminal Node to be maximized.
	//
	// The slice is provided to aid in multiplayer games
	// where multiple player results can be more simply tracked.
	//
	// It is best to return scores in the interval [-1, +1]. Using values outside this
	// range may impact search quality due to disrupting the explore-exploit tradeoff.
	// Adjust ExploreFactor proportionally if using values outside the interval.
	//
	// See github.com/wenooij/mcts/model for reusable scalar score implementations.
	Score() Score
}

type RolloutInterface interface {
	// Rollout is an optional interface method which performs random rollouts from the current node
	// and returns a raw score sum and number of rollouts performed.
	//
	// If the Search implementation does not satisfy RolloutInterface, Expand will be called during
	// the rollout instead.
	//
	// The rollout phase tends to be the most expensive part of MCTS.
	// The implementor should consider RolloutInterface if the default rollout strategy
	// proves too expensive.
	//
	// Raw score is usually defined as the sum of Scores obtained from Score after numRollouts
	// random rollouts. Note that rawScore should returned as is and not predivided by numRollouts.
	// Generally numRollouts can just be 1. numRollouts can be increased if multiple rollouts
	// per epoch is helpful.
	//
	// Backpropagation is skipped when numRollouts is 0.
	Rollout() (rawScore Score, numRollouts float64)
}
