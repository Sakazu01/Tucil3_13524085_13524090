package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"tucil3/src/algorithm"
	"tucil3/src/engine"
	"tucil3/src/utils"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Masukan file input: ")
	filePath := readLine(reader)

	puzzle, err := engine.LoadPuzzleFromFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Puzzle dimuat: %d x %d, %d checkpoint\n\n", puzzle.Rows, puzzle.Cols, puzzle.CheckpointCount())
	fmt.Println("Board Awal:")
	fmt.Print(utils.RenderBoard(puzzle, puzzle.InitialState(), nil))
	fmt.Println()

	fmt.Print("Algoritma apa yang anda pilih? (UCS/BFS/GBFS/A*/IDA*): ")
	algoStr := strings.ToUpper(readLine(reader))

	algo, ok := parseAlgorithm(algoStr)
	if !ok {
		fmt.Fprintf(os.Stderr, "Algoritma tidak valid: %q\n", algoStr)
		os.Exit(1)
	}

	var heuristic algorithm.HeuristicFunc
	heuristicName := ""
	if algorithm.AlgorithmRequiresHeuristic(algo) {
		fmt.Print("Heuristic apa yang anda pilih? (H1/H2/H3): ")
		hStr := strings.ToUpper(readLine(reader))

		heuristic, heuristicName, ok = parseHeuristic(hStr)
		if !ok {
			fmt.Fprintf(os.Stderr, "Heuristic tidak valid: %q\n", hStr)
			os.Exit(1)
		}
	}

	fmt.Println()
	if heuristicName != "" {
		fmt.Printf("Mencari solusi dengan %s + %s...\n", algoStr, heuristicName)
	} else {
		fmt.Printf("Mencari solusi dengan %s...\n", algoStr)
	}

	result, err := algorithm.Solve(puzzle, algo, heuristic)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error saat solving:", err)
		os.Exit(1)
	}

	fmt.Println()
	if !result.Found {
		fmt.Println("Tidak ada solusi yang ditemukan.")
		fmt.Printf("Waktu eksekusi        : %d ms\n", result.Duration.Milliseconds())
		fmt.Printf("Banyak iterasi        : %d iterasi\n", result.Iterations)
		return
	}

	fmt.Printf("Solusi Yang Ditemukan : %s\n", strings.Join(result.Moves, ""))
	fmt.Printf("Cost dari Solusi      : %d\n", result.TotalCost)

	fmt.Println()
	utils.PrintSolutionSteps(puzzle, result)

	fmt.Printf("Waktu eksekusi        : %d ms\n", result.Duration.Milliseconds())
	fmt.Printf("Banyak iterasi        : %d iterasi\n", result.Iterations)

	fmt.Print("\nApakah Anda ingin melakukan playback? (Ya/Tidak): ")
	if strings.EqualFold(readLine(reader), "ya") {
		fmt.Printf("Pada step berapa anda ingin melakukan playback (0-%d): ", len(result.Slides))
		startStep := 0
		if _, err := fmt.Sscanf(readLine(reader), "%d", &startStep); err != nil || startStep < 0 || startStep > len(result.Slides) {
			startStep = 0
		}
		playback(puzzle, result, reader, startStep)
	}

	fmt.Print("Apakah Anda ingin menyimpan solusi? (Ya/Tidak): ")
	if strings.EqualFold(readLine(reader), "ya") {
		fmt.Print("Masukan path file output: ")
		outPath := readLine(reader)
		if err := utils.SaveSolution(puzzle, result, algoStr, heuristicName, outPath); err != nil {
			fmt.Fprintln(os.Stderr, "Gagal menyimpan:", err)
		} else {
			fmt.Printf("Solusi disimpan pada %s\n", outPath)
		}
	}
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func parseAlgorithm(s string) (algorithm.Algorithm, bool) {
	switch s {
	case "UCS":
		return algorithm.AlgorithmUCS, true
	case "BFS":
		return algorithm.AlgorithmBFS, true
	case "GBFS":
		return algorithm.AlgorithmGBFS, true
	case "A*", "ASTAR", "A-STAR":
		return algorithm.AlgorithmAStar, true
	case "IDA*", "IDASTAR", "IDA-STAR":
		return algorithm.AlgorithmIDAStar, true
	}
	return 0, false
}

func parseHeuristic(s string) (algorithm.HeuristicFunc, string, bool) {
	switch s {
	case "H1":
		return algorithm.HeuristicManhattan, "H1 (Manhattan)", true
	case "H2":
		return algorithm.HeuristicMinimumSlide, "H2 (Minimum Slide)", true
	case "H3":
		return algorithm.HeuristicDjikstra, "H3 (Dijkstra)", true
	}
	return nil, "", false
}

func playback(puzzle *engine.Puzzle, result *algorithm.SolveResult, reader *bufio.Reader, startStep int) {
	states := buildStates(puzzle, result)
	step := startStep
	total := len(result.Slides)

	for {
		clearScreen()
		fmt.Printf("─── Playback (Langkah %d / %d) ───\n\n", step, total)

		if step == 0 {
			fmt.Println("State Awal:")
			fmt.Print(utils.RenderBoard(puzzle, puzzle.InitialState(), nil))
		} else {
			slide := result.Slides[step-1]
			fmt.Printf("Langkah %d: Geser %s  (dari %v ke %v, cost: %d)\n",
				step, slide.Direction, slide.From, slide.To, slide.PathCost)
			fmt.Print(utils.RenderBoard(puzzle, states[step], utils.BuildPathSet(slide)))
		}

		if step == total {
			fmt.Println("\nSolusi selesai!")
		}

		fmt.Printf("\n[Enter/n]=Maju  [b]=Mundur  [q]=Keluar  [angka]=Lompat ke langkah: ")
		input := strings.ToLower(readLine(reader))

		switch {
		case input == "" || input == "n":
			if step < total {
				step++
			}
		case input == "b":
			if step > 0 {
				step--
			}
		case input == "q":
			return
		default:
			var n int
			if _, err := fmt.Sscanf(input, "%d", &n); err == nil && n >= 0 && n <= total {
				step = n
			}
		}
	}
}

func buildStates(puzzle *engine.Puzzle, result *algorithm.SolveResult) []engine.State {
	states := make([]engine.State, len(result.Slides)+1)
	states[0] = puzzle.InitialState()
	for i, slide := range result.Slides {
		states[i+1] = slide.NextState
	}
	return states
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
