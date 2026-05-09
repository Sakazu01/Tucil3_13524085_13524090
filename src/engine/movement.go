package engine

import "fmt"

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

func (p *Puzzle) Slide(state State, direction string) (SlideResult, bool, error) {
	delta, exists := moveDeltas[direction]
	if !exists {
		return SlideResult{}, false, fmt.Errorf("arah gerak tidak valid: %q", direction)
	}

	current := state.Actor
	nextCheckpoint := state.NextCheckpoint
	path := make([]Position, 0, p.Rows+p.Cols)
	passedCheckpoints := make([]int, 0, 4)
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
			return p.buildSlideResult(direction, state, current, path, totalCost, passedCheckpoints, nextCheckpoint)
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
			if checkpoint == nextCheckpoint {
				passedCheckpoints = append(passedCheckpoints, checkpoint)
				nextCheckpoint++
			} else if checkpoint > nextCheckpoint {
				return SlideResult{}, false, nil
			}
		}
	}
}

func (p *Puzzle) buildSlideResult(
	direction string,
	originalState State,
	landingPos Position,
	path []Position,
	totalCost int,
	passedCheckpoints []int,
	nextCheckpoint int,
) (SlideResult, bool, error) {
	if landingPos == p.Goal && nextCheckpoint < p.CheckpointCount() {
		return SlideResult{}, false, nil
	}

	return SlideResult{
		Direction:         direction,
		From:              originalState.Actor,
		To:                landingPos,
		Path:              append([]Position(nil), path...),
		PathCost:          totalCost,
		PassedCheckpoints: append([]int(nil), passedCheckpoints...),
		NextState: State{
			Actor:          landingPos,
			NextCheckpoint: nextCheckpoint,
		},
	}, true, nil
}
