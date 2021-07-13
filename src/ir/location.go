package ir

// Location stores the position of some element, where StartChar and EndChar
// are offsets relative to the current line, as opposed to position.Position,
// where the offset is relative to the beginning of the file.
type Location struct {
	StartLine int
	EndLine   int
	StartChar int
	EndChar   int
}

func NewLocation(startLine int, endLine int, startChar int, endChar int) *Location {
	return &Location{
		StartLine: startLine,
		EndLine:   endLine,
		StartChar: startChar,
		EndChar:   endChar,
	}
}
