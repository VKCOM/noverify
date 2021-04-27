package linttest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"text/scanner"

	"github.com/VKCOM/noverify/src/cmd"
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
		errs := suite.handleFile(file)
		for _, err = range errs {
			filename := filepath.Base(file)
			t.Errorf("%s: %v", filename, err)
		}
	}
}

// handleFile processes a file from the list of php files found in the directory.
func (s *inlineTestSuite) handleFile(file string) (errs []error) {
	lines, reports, err := s.handleFileContents(file)
	if err != nil {
		return []error{err}
	}

	reportsByLine := s.createReportsByLine(reports)

	return s.handleLines(lines, reportsByLine)
}

func (s *inlineTestSuite) handleLines(lines []string, reportsByLine map[int][]string) (errs []error) {
	for index, line := range lines {
		lineIndex := index + 1
		reportByLine, hasReports := reportsByLine[lineIndex]
		expects, err := s.getExpectationForLine(line, index)
		if err != nil {
			return []error{err}
		}

		if expects == nil && hasReports {
			return []error{
				fmt.Errorf("unhandled errors for line %d: [%s]", lineIndex, strings.Join(reportByLine, ", ")),
			}
		}

		if len(expects) > 0 && !hasReports {
			return []error{
				fmt.Errorf("unexpected reports: [%s] on line %d\nexpected: no reports", strings.Join(expects, ", "), lineIndex),
			}
		}

		unmatched := s.compare(expects, reportByLine)

		for _, report := range unmatched {
			errs = append(errs, fmt.Errorf("unexpected report: '%s' on line %d\nexpected: [%s]", report, lineIndex, strings.Join(reportByLine, ", ")))
		}
	}

	return errs
}

func (s *inlineTestSuite) getExpectationForLine(line string, lineIndex int) ([]string, error) {
	commIndex := strings.Index(line, "//")
	if commIndex == -1 {
		return nil, nil
	}

	comment := line[commIndex+2:]
	p := commentParser{comment: comment, line: lineIndex}

	expects, err := p.parseExpectation()
	if err != nil {
		return nil, err
	}

	return expects, nil
}

// compare compares expected and received reports and returns a list of unmatched errors.
func (s *inlineTestSuite) compare(expects []string, reports []string) (unmatched []string) {
	for _, expect := range expects {
		var found bool

		for _, report := range reports {
			if strings.Contains(report, expect) {
				found = true
				break
			}
		}

		if !found {
			unmatched = append(unmatched, expect)
		}
	}

	return unmatched
}

// handleFileContents reads, parses the resulting file, and splits it into lines.
func (s *inlineTestSuite) handleFileContents(file string) (lines []string, reports []*linter.Report, err error) {
	lint := linter.NewLinter(linter.NewConfig())

	err = cmd.LoadEmbeddedStubs(lint, defaultStubs)
	if err != nil {
		return nil, nil, fmt.Errorf("load stubs: %v", err)
	}

	err = InitEmbeddedRules(lint.Config())
	if err != nil {
		return nil, nil, err
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("error read file '%s': %v", file, err)
	}

	content := string(data)

	// indexing
	ParseTestFile(s.t, lint, file, content)
	lint.MetaInfo().SetIndexingComplete(true)
	// analyzing
	res := ParseTestFile(s.t, lint, file, content)

	lines = strings.Split(content, "\n")

	return lines, res.Reports, nil
}

// createReportsByLine creates a map with a set of reports for each of the lines
// from the reports. This is necessary because there can be more than one report
// for one line.
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

// parseExpectation parses a string describing expected errors like
//     want `error description 1` [and` error description 2` and `error 3` ...]
func (c *commentParser) parseExpectation() (wants []string, err error) {
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

func (c *commentParser) checkKeyword(keyword string, first bool) error {
	wantKey := "and"
	if first {
		wantKey = "want"
	}

	if keyword != wantKey {
		return fmt.Errorf("expected '%s' keyword, got '%s' in '// %s', line: %d", wantKey, keyword, c.comment, c.line)
	}

	return nil
}
