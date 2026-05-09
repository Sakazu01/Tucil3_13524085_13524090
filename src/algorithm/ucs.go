package algorithm

import (
	"time"

	"tucil3/src/engine"
)

func UCS(puzzle *engine.Puzzle) (*SolveResult, error) {
	start := time.Now()
	initialState := puzzle.InitialState()

	frontier := &MinHeap{}
	frontier.Push(&SearchNode{
		State:  initialState,
		GCost:  0,
		FCost:  0,
		Moves:  []string{},
		Slides: []engine.SlideResult{},
	})

	visited := map[engine.State]bool{}
	iterations := 0

	for !frontier.IsEmpty() {
		current := frontier.Pop()

		if visited[current.State] {
			continue
		}
		visited[current.State] = true
		iterations++

		if puzzle.IsGoalState(current.State) {
			return buildResult(current, iterations, time.Since(start)), nil
		}

		for _, dir := range searchDirections {
			slide, ok, err := puzzle.Slide(current.State, dir)
			if err != nil {
				return nil, err
			}
			if !ok || visited[slide.NextState] {
				continue
			}

			nextG := current.GCost + slide.PathCost
			frontier.Push(&SearchNode{
				State:  slide.NextState,
				GCost:  nextG,
				FCost:  nextG,
				Moves:  appendMove(current.Moves, dir),
				Slides: appendSlide(current.Slides, slide),
			})
		}
	}

	return notFound(iterations, time.Since(start)), nil
}
