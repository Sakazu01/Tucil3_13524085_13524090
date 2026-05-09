package engine

import "fmt"

const (
	TileStart = "Z"
	TileGoal  = "O"
	TileWall  = "X"
	TileLava  = "L"
	TileFloor = "*"
)

var FloorTokens = map[string]struct{}{
	"*": {},
	".": {},
	"-": {},
	"_": {},
}

type Position struct {
	Row int
	Col int
}

func (p Position) Add(other Position) Position {
	return Position{Row: p.Row + other.Row, Col: p.Col + other.Col}
}

func (p Position) String() string {
	return fmt.Sprintf("(%d,%d)", p.Row, p.Col)
}

type State struct {
	Actor          Position
	NextCheckpoint int
}
