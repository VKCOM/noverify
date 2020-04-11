package scanner

type NewLines struct {
	data     []int
	src      []byte
	lastPos  int
	lastLine int
}

func NewNewLines(src []byte) NewLines {
	data := make([]int, 0, 128)
	return NewLines{src: src, data: data}
}

func (nl *NewLines) GetLine(p int) int {
	if len(nl.data) == 0 || (nl.data[len(nl.data)-1] < p && nl.lastPos < p) {
		lastLine := nl.lastLine
		for i, c := range nl.src[nl.lastPos:p] {
			if c == '\n' {
				line := nl.lastLine + i
				nl.data = append(nl.data, line)
				lastLine = line
			}
		}
		nl.lastLine = lastLine
		nl.lastPos = p
	}

	line := len(nl.data) + 1

	for i := len(nl.data) - 1; i >= 0; i-- {
		if p < nl.data[i] {
			line = i + 1
		} else {
			break
		}
	}

	return line
}
