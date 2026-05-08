package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"tucil3/src/solver"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// 1. Input file
	fmt.Print("Masukan file input: ")
	filePath := readLine(reader)

	puzzle, err := solver.LoadPuzzleFromFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Puzzle dimuat: %d x %d, %d checkpoint\n\n", puzzle.Rows, puzzle.Cols, puzzle.CheckpointCount())
	fmt.Println("Board Awal:")
	fmt.Print(renderBoard(puzzle, puzzle.InitialState(), nil))
	fmt.Println()

	// 2. Pilih algoritma
	fmt.Print("Algoritma apa yang anda pilih? (UCS/GBFS/A*): ")
	algoStr := strings.ToUpper(readLine(reader))

	algorithm, ok := parseAlgorithm(algoStr)
	if !ok {
		fmt.Fprintf(os.Stderr, "Algoritma tidak valid: %q\n", algoStr)
		os.Exit(1)
	}

	// 3. Pilih heuristik (tidak diperlukan untuk UCS)
	var heuristic solver.HeuristicFunc
	heuristicName := ""
	if algorithm != solver.AlgorithmUCS {
		fmt.Print("Heuristic apa yang anda pilih? (H1/H2/H3): ")
		hStr := strings.ToUpper(readLine(reader))

		heuristic, heuristicName, ok = parseHeuristic(hStr)
		if !ok {
			fmt.Fprintf(os.Stderr, "Heuristic tidak valid: %q\n", hStr)
			os.Exit(1)
		}
	}

	// 4. Jalankan solver
	fmt.Println()
	if heuristicName != "" {
		fmt.Printf("Mencari solusi dengan %s + %s...\n", algoStr, heuristicName)
	} else {
		fmt.Printf("Mencari solusi dengan %s...\n", algoStr)
	}

	result, err := solver.Solve(puzzle, algorithm, heuristic)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error saat solving:", err)
		os.Exit(1)
	}

	// 5. Tampilkan hasil
	fmt.Println()
	if !result.Found {
		fmt.Println("Tidak ada solusi yang ditemukan.")
	} else {
		fmt.Printf("Solusi Yang Ditemukan : %s\n", strings.Join(result.Moves, ""))
		fmt.Printf("Cost dari Solusi      : %d\n", result.TotalCost)
	}
	fmt.Printf("Waktu eksekusi        : %d ms\n", result.Duration.Milliseconds())
	fmt.Printf("Banyak iterasi        : %d iterasi\n", result.Iterations)

	if !result.Found {
		return
	}

	// Tampilkan semua langkah solusi
	fmt.Println()
	printSolutionSteps(puzzle, result)

	// 6. Tawaran playback
	fmt.Print("\nApakah Anda ingin melakukan playback? (Ya/Tidak): ")
	if strings.EqualFold(readLine(reader), "ya") {
		playback(puzzle, result, reader)
	}

	// 7. Tawaran simpan ke file
	fmt.Print("Apakah Anda ingin menyimpan solusi? (Ya/Tidak): ")
	if strings.EqualFold(readLine(reader), "ya") {
		fmt.Print("Masukan path file output: ")
		outPath := readLine(reader)
		if err := saveSolution(puzzle, result, algoStr, heuristicName, outPath); err != nil {
			fmt.Fprintln(os.Stderr, "Gagal menyimpan:", err)
		} else {
			fmt.Printf("Solusi disimpan pada %s\n", outPath)
		}
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func parseAlgorithm(s string) (solver.Algorithm, bool) {
	switch s {
	case "UCS":
		return solver.AlgorithmUCS, true
	case "GBFS":
		return solver.AlgorithmGBFS, true
	case "A*", "ASTAR", "A-STAR":
		return solver.AlgorithmAStar, true
	}
	return 0, false
}

func parseHeuristic(s string) (solver.HeuristicFunc, string, bool) {
	switch s {
	case "H1":
		return solver.HeuristicManhattan, "H1 (Manhattan)", true
	case "H2":
		return solver.HeuristicMinimumSlide, "H2 (Minimum Slide)", true
	case "H3":
		return solver.HeuristicDjikstra, "H3 (Dijkstra)", true
	}
	return nil, "", false
}

// ─── Render ─────────────────────────────────────────────────────────────────

// renderBoard merender papan puzzle sebagai string grid.
// pathSet berisi posisi lintasan slide yang ditandai '~'.
// Aktor ditampilkan sebagai '@', checkpoint yang sudah dikumpulkan sebagai '*'.
func renderBoard(puzzle *solver.Puzzle, state solver.State, pathSet map[solver.Position]bool) string {
	collectedSet := map[solver.Position]bool{}
	for i := 0; i < state.NextCheckpoint && i < len(puzzle.CheckpointOrder); i++ {
		collectedSet[puzzle.CheckpointOrder[i]] = true
	}

	var sb strings.Builder
	for row := 0; row < puzzle.Rows; row++ {
		for col := 0; col < puzzle.Cols; col++ {
			if col > 0 {
				sb.WriteByte(' ')
			}
			pos := solver.Position{Row: row, Col: col}
			switch {
			case pos == state.Actor:
				sb.WriteString("@")
			case pathSet[pos]:
				sb.WriteString("~")
			case collectedSet[pos]:
				sb.WriteString(solver.TileFloor)
			default:
				sb.WriteString(puzzle.Board[row][col])
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// buildPathSet membuat set posisi lintasan slide tanpa posisi berhenti.
func buildPathSet(slide solver.SlideResult) map[solver.Position]bool {
	set := map[solver.Position]bool{}
	for _, p := range slide.Path {
		if p != slide.To {
			set[p] = true
		}
	}
	return set
}

// ─── Output ─────────────────────────────────────────────────────────────────

func printSolutionSteps(puzzle *solver.Puzzle, result *solver.SolveResult) {
	fmt.Println("─── Langkah-langkah Solusi ───")
	fmt.Println("State Awal:")
	fmt.Print(renderBoard(puzzle, puzzle.InitialState(), nil))

	for i, slide := range result.Slides {
		fmt.Printf("\nLangkah %d: Geser %s  (dari %v ke %v, cost: %d)\n",
			i+1, slide.Direction, slide.From, slide.To, slide.PathCost)
		fmt.Print(renderBoard(puzzle, slide.NextState, buildPathSet(slide)))
	}
}

// ─── Playback ───────────────────────────────────────────────────────────────

func playback(puzzle *solver.Puzzle, result *solver.SolveResult, reader *bufio.Reader) {
	states := buildStates(puzzle, result)
	step := 0
	total := len(result.Slides)

	for {
		clearScreen()
		fmt.Printf("─── Playback (Langkah %d / %d) ───\n\n", step, total)

		if step == 0 {
			fmt.Println("State Awal:")
			fmt.Print(renderBoard(puzzle, puzzle.InitialState(), nil))
		} else {
			slide := result.Slides[step-1]
			fmt.Printf("Langkah %d: Geser %s  (dari %v ke %v, cost: %d)\n",
				step, slide.Direction, slide.From, slide.To, slide.PathCost)
			fmt.Print(renderBoard(puzzle, states[step], buildPathSet(slide)))
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
			// Coba parse sebagai nomor langkah
			var n int
			if _, err := fmt.Sscanf(input, "%d", &n); err == nil && n >= 0 && n <= total {
				step = n
			}
		}
	}
}

func buildStates(puzzle *solver.Puzzle, result *solver.SolveResult) []solver.State {
	states := make([]solver.State, len(result.Slides)+1)
	states[0] = puzzle.InitialState()
	for i, slide := range result.Slides {
		states[i+1] = slide.NextState
	}
	return states
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// ─── Save ────────────────────────────────────────────────────────────────────

func saveSolution(puzzle *solver.Puzzle, result *solver.SolveResult, algoName, heuristicName, path string) error {
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
	fmt.Fprint(w, renderBoard(puzzle, puzzle.InitialState(), nil))

	for i, slide := range result.Slides {
		fmt.Fprintf(w, "\nLangkah %d: Geser %s  (dari %v ke %v, cost: %d)\n",
			i+1, slide.Direction, slide.From, slide.To, slide.PathCost)
		fmt.Fprint(w, renderBoard(puzzle, slide.NextState, buildPathSet(slide)))
	}

	return w.Flush()
}
