package utils

import (
	"fmt"
	"strings"
	"text/scanner"
)

type CommentParser struct {
	comment string
	line    int
}

func NewCommentParser(comment string, line int) *CommentParser {
	return &CommentParser{comment: comment, line: line}
}

// ParseExpectation parses a string describing expected errors like
//     want `error description 1` [and` error description 2` and `error 3` ...]
func (c *CommentParser) ParseExpectation() (wants []string, err error) {
	// It is necessary to remove \r, since in windows the lines are separated by \r\n.
	c.comment = strings.TrimSuffix(c.comment, "\r")
	c.comment = strings.TrimLeft(c.comment, " ")
	c.comment = strings.TrimRight(c.comment, " ")

	var scanErr string
	var sc scanner.Scanner

	sc.Init(strings.NewReader(c.comment))
	sc.Mode = scanner.ScanIdents | scanner.ScanStrings | scanner.ScanRawStrings
	sc.Error = func(s *scanner.Scanner, msg string) {
		scanErr = msg + fmt.Sprintf(" in '// %s', line: %d", c.comment, c.line)
	}

	first := true

scan:
	for {
		tok := sc.Scan()

		switch tok {
		case scanner.Ident: // 'want' or 'and'
			keyword := sc.TokenText()
			if keyword != `want` && keyword != `and` {
				return nil, nil
			}

			err = c.checkKeyword(keyword, first)
			if err != nil {
				return nil, err
			}

			tok = sc.Scan()
			if tok != scanner.RawString {
				return nil, fmt.Errorf("expected value after '%s' in '// %s', line: %d", keyword, c.comment, c.line)
			}

			value := sc.TokenText()
			if len(value) <= 2 {
				return nil, fmt.Errorf("empty value after '%s' in '// %s', line: %d", keyword, c.comment, c.line)
			}

			value = value[1 : len(value)-1]

			wants = append(wants, value)
			first = false

		case scanner.EOF:
			if scanErr != "" {
				return nil, fmt.Errorf("%s", scanErr)
			}

			break scan

		default:
			return nil, fmt.Errorf("unexpected token '%s' in '// %s', line: %d", scanner.TokenString(tok), c.comment, c.line)
		}
	}

	if len(wants) == 0 {
		return nil, fmt.Errorf("empty comment on line %d", c.line)
	}

	return wants, nil
}

func (c *CommentParser) checkKeyword(keyword string, first bool) error {
	wantKey := "and"
	if first {
		wantKey = "want"
	}

	if keyword != wantKey {
		return fmt.Errorf("expected '%s' keyword, got '%s' in '// %s', line: %d", wantKey, keyword, c.comment, c.line)
	}

	return nil
}
