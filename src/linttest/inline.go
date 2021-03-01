package linttest

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
)

type inlineTestSuite struct {
	t *testing.T
}

func RunInlineTest(t *testing.T, dir string) {
	suite := inlineTestSuite{t}

	files, err := FindPHPFiles(dir)
	if err != nil {
		t.Fatalf("error find php files^ %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("source files in '%s' dir not found", dir)
	}

	for _, file := range files {
		suite.handleFile(file)
	}
}

func (s *inlineTestSuite) handleFile(file string) {
	lines, reports, err := s.handleFileContents(file)
	if err != nil {
		s.t.Error(err)
	}

	reportsByLine := s.createReportsByLine(reports)

	for line, expectedReports := range reportsByLine {
		err := s.handleReportsByLine(line, lines, expectedReports)
		if err != nil {
			s.t.Error(err)
		}
	}
}

func (s *inlineTestSuite) handleReportsByLine(line int, lines []string, expectedReports []string) error {
	if line >= len(lines) {
		return nil
	}

	lineContent := lines[line-1]
	startCommentIndex := strings.Index(lineContent, "//")

	if startCommentIndex == -1 {
		expectedReportsStr := strings.Join(expectedReports, ", ")
		return fmt.Errorf("unhandled errors for line %d: [%s]", line, expectedReportsStr)
	}

	comment := lineContent[startCommentIndex+2:] // get all after //

	p := commentParser{comment: comment, line: line}

	foundErrors, err := p.parseComment()
	if err != nil {
		return fmt.Errorf("check comments: %v", err)
	}

	for _, foundError := range foundErrors {
		var found bool

		for _, expectedMessage := range expectedReports {
			if foundError == expectedMessage {
				found = true
				break
			}
		}

		if !found {
			expectedReportsStr := strings.Join(expectedReports, ", ")
			return fmt.Errorf("unexpected report: '%s' on line %d\nexpected: [%s]", foundError, p.line, expectedReportsStr)
		}
	}
	return nil
}

func (s *inlineTestSuite) handleFileContents(file string) (lines []string, reports []*linter.Report, err error) {
	lint := linter.NewLinter(linter.NewConfig())

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("error read file '%s': %v", file, err)
	}

	content := string(data)

	// indexing
	_ = ParseTestFile(s.t, lint, file, content)
	lint.MetaInfo().SetIndexingComplete(true)
	// analyzing
	res := ParseTestFile(s.t, lint, file, content)

	lines = strings.Split(content, "\n")

	return lines, res.Reports, nil
}

func (s *inlineTestSuite) createReportsByLine(reports []*linter.Report) map[int][]string {
	reportsByLine := make(map[int][]string)

	for _, report := range reports {
		line := report.Line
		if line < 0 {
			continue
		}

		reportsByLine[line] = append(reportsByLine[line], report.Message)
	}

	return reportsByLine
}

type commentParser struct {
	comment string
	line    int
}

func (c *commentParser) parseComment() ([]string, error) {
	var wants []string

	c.comment = strings.TrimSuffix(c.comment, "\r")
	c.comment = strings.TrimLeft(c.comment, " ")
	c.comment = strings.TrimRight(c.comment, " ")

	parts := splitComment(c.comment)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty comment on line %d", c.line)
	}

	first := true
	for len(parts) != 0 {
		message, err := c.parsePartComment(parts, first)
		if err != nil {
			return nil, err
		}
		wants = append(wants, message)

		parts = parts[2:]
		first = false
	}

	return wants, nil
}

func (c *commentParser) parsePartComment(parts []string, first bool) (string, error) {
	key := parts[0]

	wantKey := "and"
	if first {
		wantKey = "want"
	}

	if key != wantKey {
		return "", fmt.Errorf("expected '%s' keyword in '// %s', line: %d", wantKey, c.comment, c.line)
	}

	if len(parts) == 1 {
		return "", fmt.Errorf("expected value after '%s' in '// %s', line: %d", wantKey, c.comment, c.line)
	}

	value := parts[1]
	if value == "" {
		return "", fmt.Errorf("empty value after '%s' in '// %s', line: %d", wantKey, c.comment, c.line)
	}

	return value, nil
}

func splitComment(comment string) []string {
	if len(comment) == 0 {
		return nil
	}

	var parts []string
	var startIndex int
	var inString bool

	for i := 0; i < len(comment); i++ {
		if comment[i] == '`' {
			if inString {
				parts = append(parts, comment[startIndex+1:i])
				startIndex = i + 1
			}

			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if comment[i] == ' ' {
			if i-startIndex > 0 {
				parts = append(parts, comment[startIndex:i])
			}
			startIndex = i + 1
		}
	}

	lastIndex := len(comment) - 1
	if lastIndex-startIndex > 0 {
		parts = append(parts, comment[startIndex:])
	}

	return parts
}
