package utils

import (
	"fmt"
	"strings"

	"tucil3/src/algorithm"
	"tucil3/src/engine"
)

func RenderBoard(puzzle *engine.Puzzle, state engine.State, pathSet map[engine.Position]bool) string {
	collectedSet := map[engine.Position]bool{}
	for i := 0; i < state.NextCheckpoint && i < len(puzzle.CheckpointOrder); i++ {
		collectedSet[puzzle.CheckpointOrder[i]] = true
	}

	var sb strings.Builder
	for row := 0; row < puzzle.Rows; row++ {
		for col := 0; col < puzzle.Cols; col++ {
			if col > 0 {
				sb.WriteByte(' ')
			}
			pos := engine.Position{Row: row, Col: col}
			switch {
			case pos == state.Actor:
				sb.WriteString("@")
			case pathSet[pos]:
				sb.WriteString("~")
			case collectedSet[pos]:
				sb.WriteString(engine.TileFloor)
			default:
				sb.WriteString(puzzle.Board[row][col])
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func BuildPathSet(slide engine.SlideResult) map[engine.Position]bool {
	set := map[engine.Position]bool{}
	for _, p := range slide.Path {
		if p != slide.To {
			set[p] = true
		}
	}
	return set
}

func PrintSolutionSteps(puzzle *engine.Puzzle, result *algorithm.SolveResult) {
	fmt.Println("─── Langkah-langkah Solusi ───")
	fmt.Println("State Awal:")
	fmt.Print(RenderBoard(puzzle, puzzle.InitialState(), nil))

	for i, slide := range result.Slides {
		fmt.Printf("\nLangkah %d: Geser %s  (dari %v ke %v, cost: %d)\n",
			i+1, slide.Direction, slide.From, slide.To, slide.PathCost)
		fmt.Print(RenderBoard(puzzle, slide.NextState, BuildPathSet(slide)))
	}
}
