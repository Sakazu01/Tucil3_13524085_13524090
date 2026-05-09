package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"tucil3/src/backend/algorithm"
	"tucil3/src/backend/engine"
)

func SaveSolution(puzzle *engine.Puzzle, result *algorithm.SolveResult, algoName, heuristicName, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	if heuristicName != "" {
		fmt.Fprintf(w, "Algoritma   : %s dengan %s\n", algoName, heuristicName)
	} else {
		fmt.Fprintf(w, "Algoritma   : %s\n", algoName)
	}
	fmt.Fprintf(w, "Solusi      : %s\n", strings.Join(result.Moves, ""))
	fmt.Fprintf(w, "Cost        : %d\n", result.TotalCost)
	fmt.Fprintf(w, "Iterasi     : %d\n", result.Iterations)
	fmt.Fprintf(w, "Waktu       : %d ms\n\n", result.Duration.Milliseconds())

	fmt.Fprintln(w, "State Awal:")
	fmt.Fprint(w, RenderBoard(puzzle, puzzle.InitialState(), nil))

	for i, slide := range result.Slides {
		fmt.Fprintf(w, "\nLangkah %d: Geser %s  (dari %v ke %v, cost: %d)\n",
			i+1, slide.Direction, slide.From, slide.To, slide.PathCost)
		fmt.Fprint(w, RenderBoard(puzzle, slide.NextState, BuildPathSet(slide)))
	}

	return w.Flush()
}
