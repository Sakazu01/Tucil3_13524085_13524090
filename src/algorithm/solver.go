package algorithm

import (
	"fmt"
	"time"

	"tucil3/src/engine"
)

type Algorithm int

const (
	AlgorithmUCS     Algorithm = iota
	AlgorithmGBFS    Algorithm = iota
	AlgorithmAStar   Algorithm = iota
	AlgorithmBFS     Algorithm = iota
	AlgorithmIDAStar Algorithm = iota
)

func (a Algorithm) String() string {
	switch a {
	case AlgorithmUCS:
		return "UCS"
	case AlgorithmGBFS:
		return "GBFS"
	case AlgorithmAStar:
		return "A*"
	case AlgorithmBFS:
		return "BFS"
	case AlgorithmIDAStar:
		return "IDA*"
	default:
		return "Unknown"
	}
}

type HeuristicFunc func(*engine.Puzzle, engine.State) (int, error)

type SolveResult struct {
	Found      bool
	Moves      []string
	Slides     []engine.SlideResult
	TotalCost  int
	Iterations int
	Duration   time.Duration
}

var searchDirections = []string{"U", "D", "L", "R"}

var algorithmsRequiringHeuristic = map[Algorithm]bool{
	AlgorithmGBFS:    true,
	AlgorithmAStar:   true,
	AlgorithmIDAStar: true,
}

func AlgorithmRequiresHeuristic(a Algorithm) bool {
	return algorithmsRequiringHeuristic[a]
}

func Solve(puzzle *engine.Puzzle, a Algorithm, heuristic HeuristicFunc) (*SolveResult, error) {
	if puzzle == nil {
		return nil, fmt.Errorf("puzzle tidak boleh nil")
	}
	if algorithmsRequiringHeuristic[a] && heuristic == nil {
		return nil, fmt.Errorf("algoritma %s memerlukan fungsi heuristik", a)
	}

	switch a {
	case AlgorithmUCS:
		return UCS(puzzle)
	case AlgorithmGBFS:
		return GBFS(puzzle, heuristic)
	case AlgorithmAStar:
		return AStar(puzzle, heuristic)
	case AlgorithmBFS:
		return BFS(puzzle)
	case AlgorithmIDAStar:
		return IDAStar(puzzle, heuristic)
	default:
		return nil, fmt.Errorf("algoritma tidak dikenal: %v", a)
	}
}

func buildResult(node *SearchNode, iterations int, duration time.Duration) *SolveResult {
	return &SolveResult{
		Found:      true,
		Moves:      node.Moves,
		Slides:     node.Slides,
		TotalCost:  node.GCost,
		Iterations: iterations,
		Duration:   duration,
	}
}

func notFound(iterations int, duration time.Duration) *SolveResult {
	return &SolveResult{Found: false, Iterations: iterations, Duration: duration}
}

func appendMove(moves []string, dir string) []string {
	next := make([]string, len(moves)+1)
	copy(next, moves)
	next[len(moves)] = dir
	return next
}

func appendSlide(slides []engine.SlideResult, slide engine.SlideResult) []engine.SlideResult {
	next := make([]engine.SlideResult, len(slides)+1)
	copy(next, slides)
	next[len(slides)] = slide
	return next
}
