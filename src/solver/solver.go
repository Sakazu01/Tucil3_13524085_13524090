package solver

import (
	"container/heap"
	"fmt"
	"time"
)

type Algorithm int

const (
	AlgorithmUCS   Algorithm = iota
	AlgorithmGBFS  Algorithm = iota
	AlgorithmAStar Algorithm = iota
)

func (a Algorithm) String() string {
	switch a {
	case AlgorithmUCS:
		return "UCS"
	case AlgorithmGBFS:
		return "GBFS"
	case AlgorithmAStar:
		return "A*"
	default:
		return "Unknown"
	}
}

// HeuristicFunc adalah signature untuk fungsi heuristik.
type HeuristicFunc func(*Puzzle, State) (int, error)

// SolveResult menyimpan solusi dan statistik pencarian.
type SolveResult struct {
	Found      bool
	Moves      []string
	Slides     []SlideResult
	TotalCost  int
	Iterations int
	Duration   time.Duration
}

// searchNode adalah node di dalam priority queue pencarian.
type searchNode struct {
	state  State
	gCost  int // biaya aktual dari start
	fCost  int // prioritas: g (UCS), h (GBFS), g+h (A*)
	moves  []string
	slides []SlideResult
	index  int
}

type searchPQ []*searchNode

func (pq searchPQ) Len() int           { return len(pq) }
func (pq searchPQ) Less(i, j int) bool { return pq[i].fCost < pq[j].fCost }
func (pq searchPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *searchPQ) Push(x any) {
	node := x.(*searchNode)
	node.index = len(*pq)
	*pq = append(*pq, node)
}
func (pq *searchPQ) Pop() any {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil
	node.index = -1
	*pq = old[:n-1]
	return node
}

var searchDirections = []string{"U", "D", "L", "R"}

// Solve menjalankan algoritma yang dipilih dan mengembalikan solusi.
// Untuk UCS, heuristic bisa nil. Untuk GBFS dan A*, heuristic harus diisi.
func Solve(puzzle *Puzzle, algorithm Algorithm, heuristic HeuristicFunc) (*SolveResult, error) {
	if puzzle == nil {
		return nil, fmt.Errorf("puzzle tidak boleh nil")
	}
	if algorithm != AlgorithmUCS && heuristic == nil {
		return nil, fmt.Errorf("GBFS dan A* memerlukan fungsi heuristik")
	}

	start := time.Now()
	initialState := puzzle.InitialState()

	h0 := 0
	if algorithm != AlgorithmUCS {
		var err error
		h0, err = heuristic(puzzle, initialState)
		if err != nil {
			return nil, fmt.Errorf("heuristic error: %w", err)
		}
	}

	initial := &searchNode{
		state:  initialState,
		gCost:  0,
		fCost:  computePriority(algorithm, 0, h0),
		moves:  []string{},
		slides: []SlideResult{},
	}

	pq := &searchPQ{initial}
	heap.Init(pq)

	// visited menyimpan state yang sudah di-expand agar tidak diproses ulang.
	visited := map[State]bool{}
	iterations := 0

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*searchNode)

		if visited[current.state] {
			continue
		}
		visited[current.state] = true
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
			result, ok, err := puzzle.Slide(current.state, dir)
			if err != nil {
				return nil, err
			}
			if !ok || visited[result.NextState] {
				continue
			}

			nextG := current.gCost + result.PathCost

			nextH := 0
			if algorithm != AlgorithmUCS {
				nextH, err = heuristic(puzzle, result.NextState)
				if err != nil {
					return nil, fmt.Errorf("heuristic error: %w", err)
				}
			}

			nextMoves := make([]string, len(current.moves)+1)
			copy(nextMoves, current.moves)
			nextMoves[len(current.moves)] = dir

			nextSlides := make([]SlideResult, len(current.slides)+1)
			copy(nextSlides, current.slides)
			nextSlides[len(current.slides)] = result

			heap.Push(pq, &searchNode{
				state:  result.NextState,
				gCost:  nextG,
				fCost:  computePriority(algorithm, nextG, nextH),
				moves:  nextMoves,
				slides: nextSlides,
			})
		}
	}

	return &SolveResult{
		Found:      false,
		Iterations: iterations,
		Duration:   time.Since(start),
	}, nil
}

func computePriority(algorithm Algorithm, g, h int) int {
	switch algorithm {
	case AlgorithmUCS:
		return g
	case AlgorithmGBFS:
		return h
	case AlgorithmAStar:
		return g + h
	default:
		return g
	}
}
