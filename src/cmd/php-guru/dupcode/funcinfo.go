package dupcode

import (
	"fmt"
)

type funcInfo struct {
	name        string
	className   string
	declPos     position
	linesOfCode int
	dups        []*funcInfo
	code        []byte
}

type position struct {
	line     int
	filename string
}

func (pos position) String() string {
	return fmt.Sprintf("%s:%d", pos.filename, pos.line)
}
