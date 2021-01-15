package quickfix

import (
	"bytes"
	"os"
	"sort"
)

// TODO: add quickfixes support in lang server?

// TextEdit is a suggested issue fix.
//
// More or less, it represents our version of the https://godoc.org/golang.org/x/tools/go/analysis#TextEdit
// which is a part of https://godoc.org/golang.org/x/tools/go/analysis#SuggestedFix.
type TextEdit struct {
	StartPos int
	EndPos   int

	// Replacement is a text to be inserted as a Replacement.
	Replacement string
}

func Apply(filename string, contents []byte, fixes []TextEdit) error {
	if len(fixes) == 0 {
		return nil
	}

	sort.Slice(fixes, func(i, j int) bool {
		return fixes[i].StartPos < fixes[j].StartPos
	})

	var buf bytes.Buffer
	buf.Grow(len(contents))
	writeFixes(&buf, contents, fixes)

	// We don't want to create a file if it doesn't exist,
	// hence using open instead of ioutil.WriteFile.
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(buf.Bytes())
	return err
}

func writeFixes(buf *bytes.Buffer, contents []byte, fixes []TextEdit) {
	offset := 0
	for _, fix := range fixes {
		// If we have a nested replacement, apply only outer replacement.
		if offset > fix.StartPos {
			continue
		}

		buf.Write(contents[offset:fix.StartPos])
		buf.WriteString(fix.Replacement)

		offset = fix.EndPos
	}
	buf.Write(contents[offset:])
}
