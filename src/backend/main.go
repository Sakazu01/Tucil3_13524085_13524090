package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tucil3/src/backend/algorithm"
	"tucil3/src/backend/engine"
)

type solveRequest struct {
	Testcase  string `json:"testcase"`
	PuzzleText string `json:"puzzleText"`
	Algorithm string `json:"algorithm"`
	Heuristic string `json:"heuristic"`
}

type apiError struct {
	Error string `json:"error"`
}

type testcaseSummary struct {
	Name string `json:"name"`
}

type testcaseResponse struct {
	Name       string    `json:"name"`
	Content    string    `json:"content"`
	Puzzle     puzzleDTO `json:"puzzle"`
	SourcePath string    `json:"sourcePath"`
}

type solveResponse struct {
	Puzzle puzzleDTO `json:"puzzle"`
	Result resultDTO `json:"result"`
	Frames []frameDTO `json:"frames"`
}

type resultDTO struct {
	Found      bool   `json:"found"`
	Algorithm  string `json:"algorithm"`
	Heuristic  string `json:"heuristic,omitempty"`
	Moves      string `json:"moves"`
	MoveCount  int    `json:"moveCount"`
	TotalCost  int    `json:"totalCost"`
	Iterations int    `json:"iterations"`
	DurationMs int64  `json:"durationMs"`
}

type puzzleDTO struct {
	Rows            int           `json:"rows"`
	Cols            int           `json:"cols"`
	Board           [][]string    `json:"board"`
	Costs           [][]int       `json:"costs"`
	Start           positionDTO   `json:"start"`
	Goal            positionDTO   `json:"goal"`
	CheckpointOrder []positionDTO `json:"checkpointOrder"`
}

type frameDTO struct {
	Step              int           `json:"step"`
	Title             string        `json:"title"`
	State             stateDTO      `json:"state"`
	Path              []positionDTO `json:"path"`
	Direction         string        `json:"direction,omitempty"`
	PathCost          int           `json:"pathCost"`
	From              positionDTO   `json:"from"`
	To                positionDTO   `json:"to"`
	PassedCheckpoints []int         `json:"passedCheckpoints"`
}

type positionDTO struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type stateDTO struct {
	Actor          positionDTO `json:"actor"`
	NextCheckpoint int         `json:"nextCheckpoint"`
}

func main() {
	frontendDir, err := resolveExistingPath(
		filepath.Join("src", "frontend"),
		filepath.Join("..", "frontend"),
	)
	if err != nil {
		log.Fatalf("gagal menemukan frontend directory: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcases", handleListTestcases)
	mux.HandleFunc("/api/testcase", handleGetTestcase)
	mux.HandleFunc("/api/solve", handleSolve)
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	addr := ":8080"
	log.Printf("Web GUI berjalan di http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

func handleListTestcases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	testDir, err := resolveTestDir()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
		return
	}

	entries, err := os.ReadDir(testDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: fmt.Sprintf("gagal membaca folder test: %v", err)})
		return
	}

	testcases := make([]testcaseSummary, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".txt") {
			continue
		}
		testcases = append(testcases, testcaseSummary{Name: entry.Name()})
	}

	sort.Slice(testcases, func(i, j int) bool {
		return testcases[i].Name < testcases[j].Name
	})

	writeJSON(w, http.StatusOK, testcases)
}

func handleGetTestcase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeJSON(w, http.StatusBadRequest, apiError{Error: "parameter name wajib diisi"})
		return
	}

	path, err := resolveTestcasePath(name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, apiError{Error: err.Error()})
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: fmt.Sprintf("gagal membaca testcase: %v", err)})
		return
	}

	puzzle, err := engine.ParsePuzzleLines(parsePuzzleText(string(content)))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: fmt.Sprintf("testcase tidak valid: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, testcaseResponse{
		Name:       name,
		Content:    string(content),
		Puzzle:     toPuzzleDTO(puzzle),
		SourcePath: path,
	})
}

func handleSolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var req solveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: fmt.Sprintf("request body tidak valid: %v", err)})
		return
	}

	puzzle, err := loadPuzzleForRequest(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	algo, err := parseAlgorithm(strings.ToUpper(strings.TrimSpace(req.Algorithm)))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	heuristicName := strings.ToUpper(strings.TrimSpace(req.Heuristic))
	heuristic, err := parseHeuristic(algo, heuristicName)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		return
	}

	result, err := algorithm.Solve(puzzle, algo, heuristic)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: fmt.Sprintf("gagal menjalankan solver: %v", err)})
		return
	}

	response := solveResponse{
		Puzzle: toPuzzleDTO(puzzle),
		Result: resultDTO{
			Found:      result.Found,
			Algorithm:  algo.String(),
			Heuristic:  heuristicName,
			Moves:      strings.Join(result.Moves, ""),
			MoveCount:  len(result.Moves),
			TotalCost:  result.TotalCost,
			Iterations: result.Iterations,
			DurationMs: result.Duration.Milliseconds(),
		},
		Frames: buildFrames(puzzle, result),
	}

	writeJSON(w, http.StatusOK, response)
}

func loadPuzzleForRequest(req solveRequest) (*engine.Puzzle, error) {
	if testcase := strings.TrimSpace(req.Testcase); testcase != "" {
		path, err := resolveTestcasePath(testcase)
		if err != nil {
			return nil, err
		}
		return engine.LoadPuzzleFromFile(path)
	}

	lines := parsePuzzleText(req.PuzzleText)
	if len(lines) == 0 {
		return nil, errors.New("isi puzzle kosong")
	}
	return engine.ParsePuzzleLines(lines)
}

func parsePuzzleText(input string) []string {
	rawLines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lines = append(lines, trimmed)
	}
	return lines
}

func parseAlgorithm(name string) (algorithm.Algorithm, error) {
	switch name {
	case "UCS":
		return algorithm.AlgorithmUCS, nil
	case "BFS":
		return algorithm.AlgorithmBFS, nil
	case "GBFS":
		return algorithm.AlgorithmGBFS, nil
	case "A*", "ASTAR", "A-STAR":
		return algorithm.AlgorithmAStar, nil
	case "IDA*", "IDASTAR", "IDA-STAR":
		return algorithm.AlgorithmIDAStar, nil
	default:
		return 0, fmt.Errorf("algoritma tidak valid: %q", name)
	}
}

func parseHeuristic(algo algorithm.Algorithm, name string) (algorithm.HeuristicFunc, error) {
	if !algorithm.AlgorithmRequiresHeuristic(algo) {
		return nil, nil
	}

	switch name {
	case "H1":
		return algorithm.HeuristicManhattan, nil
	case "H2":
		return algorithm.HeuristicMinimumSlide, nil
	case "H3":
		return algorithm.HeuristicDjikstra, nil
	default:
		return nil, fmt.Errorf("heuristic tidak valid: %q", name)
	}
}

func buildFrames(puzzle *engine.Puzzle, result *algorithm.SolveResult) []frameDTO {
	initial := puzzle.InitialState()
	frames := []frameDTO{
		{
			Step:  0,
			Title: "State Awal",
			State: toStateDTO(initial),
			From:  toPositionDTO(initial.Actor),
			To:    toPositionDTO(initial.Actor),
			Path:  []positionDTO{},
		},
	}

	for idx, slide := range result.Slides {
		frames = append(frames, frameDTO{
			Step:              idx + 1,
			Title:             fmt.Sprintf("Langkah %d", idx+1),
			State:             toStateDTO(slide.NextState),
			Path:              toPositionSliceDTO(slide.Path),
			Direction:         slide.Direction,
			PathCost:          slide.PathCost,
			From:              toPositionDTO(slide.From),
			To:                toPositionDTO(slide.To),
			PassedCheckpoints: append([]int(nil), slide.PassedCheckpoints...),
		})
	}

	return frames
}

func toPuzzleDTO(puzzle *engine.Puzzle) puzzleDTO {
	board := make([][]string, puzzle.Rows)
	costs := make([][]int, puzzle.Rows)
	for row := 0; row < puzzle.Rows; row++ {
		board[row] = append([]string(nil), puzzle.Board[row]...)
		costs[row] = append([]int(nil), puzzle.Costs[row]...)
	}

	checkpoints := make([]positionDTO, 0, len(puzzle.CheckpointOrder))
	for _, pos := range puzzle.CheckpointOrder {
		checkpoints = append(checkpoints, toPositionDTO(pos))
	}

	return puzzleDTO{
		Rows:            puzzle.Rows,
		Cols:            puzzle.Cols,
		Board:           board,
		Costs:           costs,
		Start:           toPositionDTO(puzzle.Start),
		Goal:            toPositionDTO(puzzle.Goal),
		CheckpointOrder: checkpoints,
	}
}

func toPositionDTO(pos engine.Position) positionDTO {
	return positionDTO{Row: pos.Row, Col: pos.Col}
}

func toPositionSliceDTO(path []engine.Position) []positionDTO {
	out := make([]positionDTO, 0, len(path))
	for _, pos := range path {
		out = append(out, toPositionDTO(pos))
	}
	return out
}

func toStateDTO(state engine.State) stateDTO {
	return stateDTO{
		Actor:          toPositionDTO(state.Actor),
		NextCheckpoint: state.NextCheckpoint,
	}
}

func resolveTestDir() (string, error) {
	return resolveExistingPath("test", filepath.Join("..", "..", "test"))
}

func resolveTestcasePath(name string) (string, error) {
	if strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("nama testcase tidak valid: %q", name)
	}

	testDir, err := resolveTestDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(testDir, name)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("testcase tidak ditemukan: %q", name)
	}

	return path, nil
}

func resolveExistingPath(candidates ...string) (string, error) {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("path tidak ditemukan dari kandidat: %v", candidates)
}

func writeMethodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeJSON(w, http.StatusMethodNotAllowed, apiError{Error: "method tidak diizinkan"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("gagal menulis response json: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
