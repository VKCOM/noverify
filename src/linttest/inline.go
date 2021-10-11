package linttest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/utils"
)

type inlineTestSuite struct {
	t *testing.T
}

func RunInlineTest(t *testing.T, dir string) {
	suite := inlineTestSuite{t}

	files, err := utils.FindPHPFiles(dir)
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
				fmt.Errorf("no reports matched for line %d", lineIndex),
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
	p := utils.NewCommentParser(comment, lineIndex)

	expects, err := p.ParseExpectation()
	if err != nil {
		return nil, err
	}

	return expects, nil
}

// compare expected and received reports and returns a list of unmatched errors.
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
	rawCheckerName := file
	lint := linter.NewLinter(linter.NewConfig("8.1"))
	if strings.Contains(file, "_7.4") {
		lint = linter.NewLinter(linter.NewConfig("7.4"))
		rawCheckerName = strings.ReplaceAll(file, "_7.4", "")
	}

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

	checkerName := filepath.Base(rawCheckerName)
	checkerName = checkerName[:len(checkerName)-len(filepath.Ext(file))]

	if !strings.HasSuffix(checkerName, "_any") {
		if !lint.Config().Checkers.Contains(checkerName) {
			return nil, nil, fmt.Errorf("file name must be the name of the checker that is tested. Checker %s does not exist", checkerName)
		}

		return lines, FilterReports([]string{checkerName}, res.Reports), nil
	}

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
