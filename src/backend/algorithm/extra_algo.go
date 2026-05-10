package algorithm

import (
	"time"

	"tucil3/src/backend/engine"
)

func BFS(puzzle *engine.Puzzle) (*SolveResult, error) {
	start := time.Now()
	initialState := puzzle.InitialState()

	type bfsNode struct {
		state  engine.State
		gCost  int
		moves  []string
		slides []engine.SlideResult
	}

	queue := []bfsNode{{
		state:  initialState,
		gCost:  0,
		moves:  []string{},
		slides: []engine.SlideResult{},
	}}

	visited := map[engine.State]bool{initialState: true}
	iterations := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		iterations++

		if puzzle.IsGoalState(current.state) {
			return &SolveResult{
				Found:      true,
				Moves:      current.moves,
				Slides:     current.slides,
				TotalCost:  current.gCost,
				Iterations: iterations,
				Duration:   time.Since(start),
			}, nil
		}

		for _, dir := range searchDirections {
			slide, ok, err := puzzle.Slide(current.state, dir)
			if err != nil {
				return nil, err
			}
			if !ok || visited[slide.NextState] {
				continue
			}

			visited[slide.NextState] = true
			queue = append(queue, bfsNode{
				state:  slide.NextState,
				gCost:  current.gCost + slide.PathCost,
				moves:  appendMove(current.moves, dir),
				slides: appendSlide(current.slides, slide),
			})
		}
	}

	return notFound(iterations, time.Since(start)), nil
}

func IDAStar(puzzle *engine.Puzzle, heuristic HeuristicFunc) (*SolveResult, error) {
	start := time.Now()
	initialState := puzzle.InitialState()

	h0, err := heuristic(puzzle, initialState)
	if err != nil {
		return nil, err
	}

	threshold := h0
	totalIterations := 0

	for {
		pathVisited := map[engine.State]bool{initialState: true}

		result, nextThreshold, iterCount, err := idaStarDFS(
			puzzle, initialState, 0, threshold, heuristic,
			pathVisited, []string{}, []engine.SlideResult{},
		)
		totalIterations += iterCount

		if err != nil {
			return nil, err
		}
		if result != nil {
			result.Iterations = totalIterations
			result.Duration = time.Since(start)
			return result, nil
		}
		if nextThreshold >= unreachableHeuristic {
			return notFound(totalIterations, time.Since(start)), nil
		}

		threshold = nextThreshold
	}
}

func idaStarDFS(
	puzzle *engine.Puzzle,
	state engine.State,
	gCost int,
	threshold int,
	heuristic HeuristicFunc,
	pathVisited map[engine.State]bool,
	moves []string,
	slides []engine.SlideResult,
) (*SolveResult, int, int, error) {
	h, err := heuristic(puzzle, state)
	if err != nil {
		return nil, 0, 0, err
	}

	fCost := gCost + h
	if fCost > threshold {
		return nil, fCost, 0, nil
	}

	if puzzle.IsGoalState(state) {
		return &SolveResult{
			Found:     true,
			Moves:     moves,
			Slides:    slides,
			TotalCost: gCost,
		}, 0, 1, nil
	}

	minNext := unreachableHeuristic
	iterCount := 1

	for _, dir := range searchDirections {
		slide, ok, err := puzzle.Slide(state, dir)
		if err != nil {
			return nil, 0, iterCount, err
		}
		if !ok || pathVisited[slide.NextState] {
			continue
		}

		pathVisited[slide.NextState] = true

		childResult, childNext, childIter, err := idaStarDFS(
			puzzle, slide.NextState, gCost+slide.PathCost, threshold, heuristic,
			pathVisited, appendMove(moves, dir), appendSlide(slides, slide),
		)
		iterCount += childIter
		pathVisited[slide.NextState] = false

		if err != nil {
			return nil, 0, iterCount, err
		}
		if childResult != nil {
			return childResult, 0, iterCount, nil
		}
		if childNext < minNext {
			minNext = childNext
		}
	}

	return nil, minNext, iterCount, nil
}
