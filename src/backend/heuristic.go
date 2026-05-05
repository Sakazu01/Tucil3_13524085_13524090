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

type Heuristic struct {
	Name string
}

func NewHeuristic(name string) *Heuristic {
	_ = bufio.NewScanner
	_ = heap.Init
	_ = fmt.Sprintf
	_ = os.Open
	_ = strconv.Atoi
	_ = strings.TrimSpace
	_ = time.Second

	return &Heuristic{Name: name}
}

func (h *Heuristic) Evaluate(puzzle *Puzzle, state State) (int, error) {
	if puzzle == nil {
		return 0, fmt.Errorf("puzzle tidak boleh nil")
	}

	_ = state
	return 0, fmt.Errorf("heuristic %q belum diimplementasikan", h.Name)
}
