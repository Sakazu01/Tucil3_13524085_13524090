package engine

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func LoadPuzzleFromFile(path string) (*Puzzle, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("gagal membaca file: %w", err)
	}

	return ParsePuzzleLines(lines)
}

func ParsePuzzleLines(lines []string) (*Puzzle, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("input puzzle kosong")
	}

	rows, cols, err := parseDimensions(lines[0])
	if err != nil {
		return nil, err
	}

	expected := 1 + rows + rows
	if len(lines) != expected {
		return nil, fmt.Errorf("jumlah baris input tidak valid: expected %d, got %d", expected, len(lines))
	}

	board := make([][]string, rows)
	for row := 0; row < rows; row++ {
		board[row], err = parseBoardLine(lines[1+row], cols)
		if err != nil {
			return nil, fmt.Errorf("board baris %d: %w", row+1, err)
		}
	}

	costs := make([][]int, rows)
	for row := 0; row < rows; row++ {
		costs[row], err = parseCostLine(lines[1+rows+row], cols)
		if err != nil {
			return nil, fmt.Errorf("cost baris %d: %w", row+1, err)
		}
	}

	puzzle := &Puzzle{
		Rows:            rows,
		Cols:            cols,
		Board:           board,
		Costs:           costs,
		Checkpoints:     make(map[int]Position),
		MaxCheckpoint:   -1,
		MinWalkableCost: -1,
	}

	if err := puzzle.validateAndIndex(); err != nil {
		return nil, err
	}

	return puzzle, nil
}

func (p *Puzzle) validateAndIndex() error {
	startCount, goalCount := 0, 0

	for row := 0; row < p.Rows; row++ {
		if len(p.Board[row]) != p.Cols {
			return fmt.Errorf("panjang board baris %d tidak konsisten", row+1)
		}
		if len(p.Costs[row]) != p.Cols {
			return fmt.Errorf("panjang cost baris %d tidak konsisten", row+1)
		}

		for col := 0; col < p.Cols; col++ {
			pos := Position{Row: row, Col: col}
			token := p.Board[row][col]
			cost := p.Costs[row][col]

			if cost < 0 {
				return fmt.Errorf("cost tidak boleh negatif di %s", pos)
			}

			switch token {
			case TileStart:
				startCount++
				p.Start = pos
			case TileGoal:
				goalCount++
				p.Goal = pos
			}

			if idx, ok := parseCheckpointToken(token); ok {
				if _, exists := p.Checkpoints[idx]; exists {
					return fmt.Errorf("checkpoint %d muncul lebih dari sekali", idx)
				}
				p.Checkpoints[idx] = pos
				if idx > p.MaxCheckpoint {
					p.MaxCheckpoint = idx
				}
			}

			if p.IsWalkable(token) && (p.MinWalkableCost == -1 || cost < p.MinWalkableCost) {
				p.MinWalkableCost = cost
			}
		}
	}

	if startCount != 1 {
		return fmt.Errorf("jumlah tile Z harus tepat 1, ditemukan %d", startCount)
	}
	if goalCount != 1 {
		return fmt.Errorf("jumlah tile O harus tepat 1, ditemukan %d", goalCount)
	}
	if p.MinWalkableCost < 0 {
		return fmt.Errorf("tidak ada tile walkable yang valid")
	}

	p.CheckpointOrder = make([]Position, 0)
	for idx := 0; idx <= p.MaxCheckpoint; idx++ {
		pos, exists := p.Checkpoints[idx]
		if !exists {
			return fmt.Errorf("urutan checkpoint tidak lengkap: checkpoint %d tidak ditemukan", idx)
		}
		p.CheckpointOrder = append(p.CheckpointOrder, pos)
	}

	return nil
}

func parseDimensions(line string) (int, int, error) {
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("baris pertama harus berformat: N M")
	}
	rows, err := strconv.Atoi(fields[0])
	if err != nil || rows <= 0 {
		return 0, 0, fmt.Errorf("N tidak valid: %q", fields[0])
	}
	cols, err := strconv.Atoi(fields[1])
	if err != nil || cols <= 0 {
		return 0, 0, fmt.Errorf("M tidak valid: %q", fields[1])
	}
	return rows, cols, nil
}

func parseBoardLine(line string, cols int) ([]string, error) {
	fields := strings.Fields(line)
	var row []string

	switch {
	case len(fields) == cols:
		row = append(row, fields...)
	case len(fields) == 1:
		for _, r := range fields[0] {
			row = append(row, string(r))
		}
	default:
		return nil, fmt.Errorf("jumlah token board harus %d atau ditulis compact tanpa spasi", cols)
	}

	if len(row) != cols {
		return nil, fmt.Errorf("jumlah kolom board harus %d, ditemukan %d", cols, len(row))
	}

	for _, token := range row {
		if !isAllowedBoardToken(token) {
			return nil, fmt.Errorf("token board tidak valid: %q", token)
		}
	}

	return row, nil
}

func parseCostLine(line string, cols int) ([]int, error) {
	fields := strings.Fields(line)
	if len(fields) != cols {
		return nil, fmt.Errorf("jumlah cost harus %d, ditemukan %d", cols, len(fields))
	}

	row := make([]int, cols)
	for i, field := range fields {
		v, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("cost pada kolom %d tidak valid: %q", i+1, field)
		}
		if v < 0 {
			return nil, fmt.Errorf("cost pada kolom %d tidak boleh negatif", i+1)
		}
		row[i] = v
	}

	return row, nil
}

func isAllowedBoardToken(token string) bool {
	switch token {
	case TileStart, TileGoal, TileWall, TileLava:
		return true
	}
	if _, ok := FloorTokens[token]; ok {
		return true
	}
	_, ok := parseCheckpointToken(token)
	return ok
}

func parseCheckpointToken(token string) (int, bool) {
	if token == "" {
		return 0, false
	}
	for _, c := range token {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	if len(token) > 1 && token[0] == '0' {
		return 0, false
	}
	v, err := strconv.Atoi(token)
	if err != nil || v < 0 {
		return 0, false
	}
	return v, true
}
