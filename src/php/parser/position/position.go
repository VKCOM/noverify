package position

import (
	"fmt"
)

// Position represents node position
type Position struct {
	StartLine int
	EndLine   int
	StartPos  int
	EndPos    int
}

// NewPosition Position constructor
func NewPosition(startLine int, endLine int, startPos int, endPos int) *Position {
	return &Position{
		StartLine: startLine,
		EndLine:   endLine,
		StartPos:  startPos,
		EndPos:    endPos,
	}
}

func (p Position) String() string {
	return fmt.Sprintf("Pos{Line: %d-%d Pos: %d-%d}", p.StartLine, p.EndLine, p.StartPos, p.EndPos)
}
