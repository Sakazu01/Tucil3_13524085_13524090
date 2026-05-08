package solver

import (
	"fmt"
	"strings"
)

var moveDeltas = map[string]Position{
	"U": {Row: -1, Col: 0},
	"D": {Row: 1, Col: 0},
	"L": {Row: 0, Col: -1},
	"R": {Row: 0, Col: 1},
}

type SlideResult struct {
	Direction         string
	From              Position
	To                Position
	Path              []Position
	PathCost          int
	PassedCheckpoints []int
	NextState         State
}

func (p *Puzzle) InitialState() State {
	return State{
		Actor:          p.Start,
		NextCheckpoint: 0,
	}
}

func (p *Puzzle) CheckpointCount() int {
	return len(p.CheckpointOrder)
}

func (p *Puzzle) InBounds(pos Position) bool {
	return pos.Row >= 0 && pos.Row < p.Rows && pos.Col >= 0 && pos.Col < p.Cols
}

func (p *Puzzle) Cell(pos Position) string {
	return p.Board[pos.Row][pos.Col]
}

func (p *Puzzle) CostAt(pos Position) int {
	return p.Costs[pos.Row][pos.Col]
}

func (p *Puzzle) IsWalkableToken(token string) bool {
	return token != TileWall && token != TileLava
}

func (p *Puzzle) IsGoalState(state State) bool {
	return state.Actor == p.Goal && state.NextCheckpoint == p.CheckpointCount()
}

func (p *Puzzle) NextTarget(state State) Position {
	if state.NextCheckpoint < p.CheckpointCount() {
		return p.CheckpointOrder[state.NextCheckpoint]
	}
	return p.Goal
}

func (p *Puzzle) Slide(state State, direction string) (SlideResult, bool, error) {
	delta, exists := moveDeltas[strings.ToUpper(strings.TrimSpace(direction))]
	if !exists {
		return SlideResult{}, false, fmt.Errorf("arah gerak tidak valid: %q", direction)
	}

	current := state.Actor
	nextCheckpoint := state.NextCheckpoint
	path := make([]Position, 0, p.Rows+p.Cols)
	passedCheckpoints := make([]int, 0, 2)
	totalCost := 0
	moved := false

	for {
		next := current.Add(delta)
		if !p.InBounds(next) {
			return SlideResult{}, false, nil
		}

		token := p.Cell(next)
		if token == TileWall {
			if !moved {
				return SlideResult{}, false, nil
			}

			if current == p.Goal && nextCheckpoint < p.CheckpointCount() {
				return SlideResult{}, false, nil
			}

			return SlideResult{
				Direction: direction,
				From:      state.Actor,
				To:        current,
				Path:      append([]Position(nil), path...),
				PathCost:  totalCost,
				PassedCheckpoints: append(
					[]int(nil),
					passedCheckpoints...,
				),
				NextState: State{
					Actor:          current,
					NextCheckpoint: nextCheckpoint,
				},
			}, true, nil
		}

		if token == TileLava {
			return SlideResult{}, false, nil
		}

		current = next
		moved = true
		path = append(path, current)
		totalCost += p.CostAt(current)

		checkpoint, isCheckpoint := parseCheckpointToken(token)
		if isCheckpoint {
			switch {
			case checkpoint < nextCheckpoint:
			case checkpoint == nextCheckpoint:
				passedCheckpoints = append(passedCheckpoints, checkpoint)
				nextCheckpoint++
			default:
				return SlideResult{}, false, nil
			}
		}
	}
}
