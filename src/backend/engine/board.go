package engine

type Puzzle struct {
	Rows            int
	Cols            int
	Board           [][]string
	Costs           [][]int
	Start           Position
	Goal            Position
	Checkpoints     map[int]Position
	CheckpointOrder []Position
	MaxCheckpoint   int
	MinWalkableCost int
}

func (p *Puzzle) InitialState() State {
	return State{Actor: p.Start, NextCheckpoint: 0}
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

func (p *Puzzle) IsWalkable(token string) bool {
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
