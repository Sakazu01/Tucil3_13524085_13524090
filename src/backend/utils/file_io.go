package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"tucil3/src/backend/algorithm"
	"tucil3/src/backend/engine"
)

func BuildSolutionText(w io.Writer, puzzle *engine.Puzzle, result *algorithm.SolveResult, testcaseName, algoName, heuristicName string) error {
	bw := bufio.NewWriter(w)

	algoLabel := algoName
	if heuristicName != "" {
		algoLabel = fmt.Sprintf("%s (%s)", algoName, heuristicName)
	}

	fmt.Fprintf(bw, "Testcase  : %s\n", testcaseName)
	fmt.Fprintf(bw, "Algorithm : %s\n", algoLabel)
	fmt.Fprintf(bw, "Heuristik : %s\n", heuristicName)
	fmt.Fprintf(bw, "Found     : %v\n", result.Found)
	fmt.Fprintf(bw, "Moves     : %s\n", strings.Join(result.Moves, ""))
	fmt.Fprintf(bw, "MoveCount : %d\n", len(result.Moves))
	fmt.Fprintf(bw, "TotalCost : %d\n", result.TotalCost)
	fmt.Fprintf(bw, "Iterations: %d\n", result.Iterations)
	fmt.Fprintf(bw, "Duration  : %d ms\n", result.Duration.Milliseconds())

	return bw.Flush()
}

func SaveSolution(puzzle *engine.Puzzle, result *algorithm.SolveResult, testcaseName, algoName, heuristicName, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return BuildSolutionText(f, puzzle, result, testcaseName, algoName, heuristicName)
}
