// Package mcts provides an implementation of general multi-agent Monte-Carlo tree search (MCTS).
package mcts

import "fmt"

// Step represents an atomic an unambiguous transition in a game tree.
//
// String should return the Step's standard unambiguous representation.
type Step interface {
	comparable
	fmt.Stringer
}

// FrontierStep wraps a step returned from Expand with extra parameters to apply to its subtree.
//
// Priority is used to seed the initial ordering of steps in the MAB min-priority-queue.
// Ideally, the priority value for X should be set to -E[Score(X)], but in practice
// a heuristic is used. Like ExploreFactor, it is critical that the priority be roughly
// proportional to the values returned from Score, otherwise a small value such as -∞
// can be used to guarantee that expanded nodes are explored at least once.
// In larger state spaces this may be counterproductive.
// Priority only affects the initial value. The next priority is recomputed in backprop.
//
// ExploreFactor defines the exploration weighting for the node and its subtree.
// If this is 0, the parent's ExploreFactor is copied. By default, a ExploreFactor is
// applied uniformly to all nodes. It is critical that ExploreFactor be roughly
// proportional to the values returned from Score.
type FrontierStep[S Step] struct {
	Step          S
	InitialScore  Score
	Priority      float64
	ExploreFactor float64
}

// SearchInterface is the minimal interface to MCTS tree state.
//
// SearchInterface provides a simple flat interface:
//
//   - Select moves the state of the game forward one Step.
//   - Expand reveals available Steps at the current state.
//   - Root reset the game tree to the state at the beginning of search.
//
// The game tree is managed by this package, so it is not required to implement
// one yourself, so long as Select and Root are relatively inexpensive.
// Expand will not be called again for nodes in the game tree.
//
// MCTS proceeds from the root, down the game tree, selecting the best Step
// at each level, until it reaches a frontier leaf node where a random rollout
// is started. The implementor has two choices for the rollout:
//
//  1. Implement ScoreInterface and rely on the defeault rollout strategy.
//  2. Implement RolloutInterface and use a custom strategy.
//
// Which one to use will depend on one's needs:
//
// RolloutInterface allows for a custom implementation of rollouts.
// This may be useful for those who wish to apply custom heuristics for sampling steps.
// Additionally, this may be a good choice when Expand is expensive.
// Lastly, RolloutInterface supports multiple simulations per epoch
// which may provide further flexibility. For instance, running extra
// simulations for promising positions.
//
// Otherwise, the default rollout strategy can be used. This uses the base methods from
// SearchInterface, calling Select and Expand repeatedly until a terminal state has been
// reached (Expand returns an empty slice of steps).
// The resulting score from ScoreInterface.Score is returned and the rollout completes.
// If the implementation does not satisfy ScoreInterface, a random normally distributed
// dummy score is returned.
// The default rollout strategy may be good enough for one's needs and one done not have
// to bother implementing a custom strategy.
// Alternative implementations of Expand which apply heuristics or generates only a single
// steps must be carefully coded. For instance, failure to return available steps randomly
// will introduce biases to search. It has also been reported that optimal play in simulation
// may actually hinder the explorative performance of MCTS.
// Note that Expand can be made less expensive by reusing the same slice.
// The slice will not be retained by the implementation in this package.
type SearchInterface[S Step] interface {
	// Root resets the current search to root.
	//
	// Root is called multiple times in Search before the selection phase
	// and after Search completes.
	Root()

	// Select the Step in the current node.
	//
	// Select is called multiple times during the selection phase.
	// Select will also be called during rollout if Search does not implement RolloutInterface.
	Select(S)

	// Expand returns all steps in the current state.
	//
	// Expand is called after the selection phase to expand the frontier of a leaf node.
	// If Expand returns no steps, the node is marked as a terminal.
	//
	// Expand will also be called during rollout if Search does not implement RolloutInterface.
	// Expand must always eventually return a terminal if using the default rollout strategy.
	Expand() []FrontierStep[S]
}

// Score is an interface for scorekeeping in search.
//
// Score is provided to simplify the implementation of
// multiplayer games where multiple scores will be tracked at each position.
//
// For single-player applications implementations can wrap a single float64.
//
// See github.com/wenooij/mcts/model for reusable scalar score implementations.
type Score interface {
	// Score returns the objective evaluation for the Score.
	//
	// The Score value will be maximized so Score should be relative to
	// the current position and side to move in multiplayer contexts.
	//
	// It is best to return scores in the interval [-1, +1]. Using values outside this
	// range may impact search quality due to disrupting the explore-exploit tradeoff.
	// Adjust ExploreFactor proportionally if using values outside the interval.
	Score() float64

	// Add returns the sum of Scores of the same type.
	//
	// Add may have a pointer receiver and modify itself so long as it returns the result.
	Add(Score) Score
}

type ScoreInterface interface {
	// Score is an optional interface which returns an objective evaluation in
	// terminal positions.
	//
	// Score will be exactly once in for every state selected.
	// This ensures the score can be propogated through the tree without any nils.
	// Score should take care to return a score with the correct parity in games
	// with multiple players.
	//
	// Implementors may chose to implement RolloutInterface instead of Score
	// if custom Rollout implementations are being used.
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
	Rollout() (rawScore Score, numRollouts int)
}
