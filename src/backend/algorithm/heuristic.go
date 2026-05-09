package algorithm

import (
	"container/heap"
	"fmt"

	"tucil3/src/backend/engine"
)

var heuristicDirections = []string{"U", "D", "L", "R"}

const unreachableHeuristic = int(^uint(0) >> 2)

var heuristicDjikstraCache = map[*engine.Puzzle]map[engine.State]int{}

func HeuristicManhattan(puzzle *engine.Puzzle, state engine.State) (int, error) {
	if err := validateHeuristicInput(puzzle, state); err != nil {
		return 0, err
	}

	current := state.Actor
	total := 0

	for idx := state.NextCheckpoint; idx < puzzle.CheckpointCount(); idx++ {
		target := puzzle.CheckpointOrder[idx]
		total += manhattanDistance(current, target)
		current = target
	}

	total += manhattanDistance(current, puzzle.Goal)
	return total * puzzle.MinWalkableCost, nil
}

func HeuristicMinimumSlide(puzzle *engine.Puzzle, state engine.State) (int, error) {
	if err := validateHeuristicInput(puzzle, state); err != nil {
		return 0, err
	}
	if targetSatisfied(puzzle, state, state) {
		return 0, nil
	}

	type bfsNode struct {
		state engine.State
		depth int
	}

	queue := []bfsNode{{state: state, depth: 0}}
	visited := map[engine.State]struct{}{state: {}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, direction := range heuristicDirections {
			result, ok, err := puzzle.Slide(current.state, direction)
			if err != nil {
				return 0, err
			}
			if !ok {
				continue
			}

			nextState := result.NextState
			if targetSatisfied(puzzle, state, nextState) {
				return (current.depth + 1) * puzzle.MinWalkableCost, nil
			}

			if _, exists := visited[nextState]; exists {
				continue
			}

			visited[nextState] = struct{}{}
			queue = append(queue, bfsNode{state: nextState, depth: current.depth + 1})
		}
	}

	return unreachableHeuristic, nil
}

func HeuristicDjikstra(puzzle *engine.Puzzle, state engine.State) (int, error) {
	if err := validateHeuristicInput(puzzle, state); err != nil {
		return 0, err
	}

	cache, exists := heuristicDjikstraCache[puzzle]
	if !exists {
		cache = make(map[engine.State]int)
		heuristicDjikstraCache[puzzle] = cache
	}

	if value, exists := cache[state]; exists {
		return value, nil
	}
	if targetSatisfied(puzzle, state, state) {
		cache[state] = 0
		return 0, nil
	}

	distances := map[engine.State]int{state: 0}

	pq := &statePriorityQueue{{state: state, cost: 0}}
	heap.Init(pq)

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*statePriorityItem)
		bestCost, exists := distances[current.state]
		if !exists || current.cost != bestCost {
			continue
		}

		if targetSatisfied(puzzle, state, current.state) {
			cache[state] = current.cost
			return current.cost, nil
		}

		for _, direction := range heuristicDirections {
			result, ok, err := puzzle.Slide(current.state, direction)
			if err != nil {
				return 0, err
			}
			if !ok {
				continue
			}

			nextCost := current.cost + result.PathCost
			prevCost, seen := distances[result.NextState]
			if seen && nextCost >= prevCost {
				continue
			}

			distances[result.NextState] = nextCost
			heap.Push(pq, &statePriorityItem{state: result.NextState, cost: nextCost})
		}
	}

	cache[state] = unreachableHeuristic
	return unreachableHeuristic, nil
}

func validateHeuristicInput(puzzle *engine.Puzzle, state engine.State) error {
	if puzzle == nil {
		return fmt.Errorf("puzzle tidak boleh nil")
	}
	if state.NextCheckpoint < 0 || state.NextCheckpoint > puzzle.CheckpointCount() {
		return fmt.Errorf("next checkpoint tidak valid: %d", state.NextCheckpoint)
	}
	if !puzzle.InBounds(state.Actor) {
		return fmt.Errorf("posisi aktor di luar papan: %s", state.Actor)
	}
	return nil
}

func targetSatisfied(puzzle *engine.Puzzle, initial, current engine.State) bool {
	if initial.NextCheckpoint < puzzle.CheckpointCount() {
		return current.NextCheckpoint > initial.NextCheckpoint
	}
	return puzzle.IsGoalState(current)
}

func manhattanDistance(a, b engine.Position) int {
	return abs(a.Row-b.Row) + abs(a.Col-b.Col)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

type statePriorityItem struct {
	state engine.State
	cost  int
	index int
}

type statePriorityQueue []*statePriorityItem

func (pq statePriorityQueue) Len() int            { return len(pq) }
func (pq statePriorityQueue) Less(i, j int) bool  { return pq[i].cost < pq[j].cost }
func (pq statePriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *statePriorityQueue) Push(item any) {
	node := item.(*statePriorityItem)
	node.index = len(*pq)
	*pq = append(*pq, node)
}

func (pq *statePriorityQueue) Pop() any {
	old := *pq
	last := len(old) - 1
	item := old[last]
	old[last] = nil
	item.index = -1
	*pq = old[:last]
	return item
}
