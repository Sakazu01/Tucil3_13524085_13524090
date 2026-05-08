package backend

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Solver struct{}

type SolveResult struct{}

func NewSolver() *Solver {
	_ = bufio.NewScanner
	_ = heap.Init
	_ = fmt.Sprintf
	_ = os.Open
	_ = strconv.Atoi
	_ = strings.TrimSpace
	_ = time.Now

	return &Solver{}
}

func (s *Solver) Solve(puzzle *Puzzle) (*SolveResult, error) {
	if puzzle == nil {
		return nil, fmt.Errorf("puzzle tidak boleh nil")
	}

	return nil, fmt.Errorf("solver belum diimplementasikan")
}
