package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/utils"
	"github.com/VKCOM/noverify/src/workspace"
)

func RegisterTestRulesFlags(ctx *AppContext) (*flag.FlagSet, *FlagsGroups) {
	flags := &RulesTestSuite{}
	fs := flag.NewFlagSet("test-rules", flag.ContinueOnError)
	groups := NewFlagsGroups()

	groups.AddGroup("Files")
	groups.AddGroup("Other")

	fs.StringVar(&flags.rules, "rules", "./", "Comma separated list of directories or files with rules")
	fs.StringVar(&flags.tests, "tests", "./tests", "Directory or file with tests")

	groups.Add("Files", "rules")
	groups.Add("Files", "tests")

	fs.BoolVar(&flags.kphp, "kphp", false, "KPHP mode")

	groups.Add("Other", "kphp")

	ctx.CustomFlags = flags
	return fs, groups
}

func TestRules(ctx *AppContext) (status int, err error) {
	log.SetFlags(0)
	flags := ctx.CustomFlags.(*RulesTestSuite)

	suite := NewRulesTestSuite(flags.rules, flags.tests, flags.kphp)
	err = suite.Run()
	if err != nil {
		status = 2
	}

	if err == nil {
		log.Println("[Tests Passed]")
	} else {
		log.Println("[Tests Failed]")
	}

	return status, nil
}

type RulesTestSuite struct {
	rules string
	tests string

	kphp bool
}

func NewRulesTestSuite(rules string, tests string, kphp bool) *RulesTestSuite {
	return &RulesTestSuite{rules: rules, tests: tests, kphp: kphp}
}

func (s *RulesTestSuite) Run() error {
	go linter.MemoryLimiterThread(0)

	files, err := utils.FindPHPFiles(s.tests)
	if err != nil {
		return fmt.Errorf("error find php files %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("tests files in '%s' not found", s.tests)
	}

	failed := false

	for _, file := range files {
		errs := s.handleFile(file)
		if len(errs) > 0 {
			log.Printf("Test '%s' failed\n", file)
			for _, err := range errs {
				// filename := filepath.Base(file)
				log.Printf("   %v\n", err)
			}

			failed = true
		} else {
			log.Printf("Test '%s' passed\n", file)
		}
	}

	if failed {
		return fmt.Errorf("test failed")
	}
	return nil
}

// handleFile processes a file from the list of php files found in the directory.
func (s *RulesTestSuite) handleFile(file string) (errs []error) {
	lines, reports, err := s.handleFileContents(file)
	if err != nil {
		return []error{err}
	}

	reportsByLine := s.createReportsByLine(reports)

	return s.handleLines(lines, reportsByLine)
}

func (s *RulesTestSuite) handleLines(lines []string, reportsByLine map[int][]string) (errs []error) {
	for index, line := range lines {
		lineIndex := index + 1
		reportByLine, hasReports := reportsByLine[lineIndex]
		expects, err := s.getExpectationForLine(line, index)
		if err != nil {
			return []error{err}
		}

		if expects == nil && hasReports {
			for i := range reportByLine {
				reportByLine[i] = "'" + reportByLine[i] + "'"
			}
			return []error{
				fmt.Errorf("unhandled errors for line %d, expected: [%s]", lineIndex, strings.Join(reportByLine, ", ")),
			}
		}

		if len(expects) > 0 && !hasReports {
			return []error{
				fmt.Errorf("no reports matched for line %d", lineIndex),
			}
		}

		unmatched := s.compare(expects, reportByLine)

		for _, report := range unmatched {
			for i := range reportByLine {
				reportByLine[i] = "'" + reportByLine[i] + "'"
			}
			errs = append(errs, fmt.Errorf("unexpected report: '%s' on line %d, expected: [%s]", report, lineIndex, strings.Join(reportByLine, ", ")))
		}
	}

	return errs
}

func (s *RulesTestSuite) getExpectationForLine(line string, lineIndex int) ([]string, error) {
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
func (s *RulesTestSuite) compare(expects []string, reports []string) (unmatched []string) {
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
func (s *RulesTestSuite) handleFileContents(file string) (lines []string, reports []*linter.Report, err error) {
	rawCheckerName := file
	lint := linter.NewLinter(linter.NewConfig("8.1"))
	if strings.Contains(file, "_7.4") {
		lint = linter.NewLinter(linter.NewConfig("7.4"))
		rawCheckerName = strings.ReplaceAll(file, "_7.4", "")
	}

	runner := NewLinterRunner(lint, linter.NewCheckersFilterWithEnabledAll())

	err = InitStubs(lint)
	if err != nil {
		return nil, nil, fmt.Errorf("load stubs: %v", err)
	}

	err = s.initEmbeddedRules(lint.Config())
	if err != nil {
		return nil, nil, err
	}

	config := lint.Config()
	config.KPHP = s.kphp

	ruleSets, err := ParseExternalRules(s.rules)
	if err != nil {
		return nil, nil, fmt.Errorf("preload external rules: %v", err)
	}

	for _, rset := range ruleSets {
		config.Checkers.DeclareRules(rset)
	}

	err = runner.Init(ruleSets, &ParsedFlags{
		AllowAll:         true,
		AllowChecks:      AllChecks,
		PhpExtensionsArg: "php,inc,php5,phtml",
		MaxConcurrency:   runtime.NumCPU(),
		MaxFileSize:      20 * 1024 * 1024,
		UnusedVarPattern: "^_$",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("runner init fail: %v", err)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("error read file '%s': %v", file, err)
	}

	content := string(data)

	// indexing
	s.ParseTestFile(lint, file, content)
	lint.MetaInfo().SetIndexingComplete(true)
	// analyzing
	res := s.ParseTestFile(lint, file, content)

	checkerNamespace := ""
	for _, stmt := range res.RootNode.Stmts {
		if namespace, ok := stmt.(*ir.NamespaceStmt); ok {
			checkerNamespace = strings.ReplaceAll(namespace.NamespaceName.Value, `\`, `/`) + `/`
			break
		}
	}

	lines = strings.Split(content, "\n")

	checkerName := filepath.Base(rawCheckerName)
	checkerName = checkerName[:len(checkerName)-len(filepath.Ext(file))]
	checkerName = checkerNamespace + checkerName

	if !strings.HasSuffix(checkerName, "_any") {
		if !lint.Config().Checkers.Contains(checkerName) {
			return nil, nil, fmt.Errorf("file name with namespace inside must be the name of the checker that is tested. Checker '%s' does not exist", checkerName)
		}

		return lines, s.filterReports([]string{checkerName}, res.Reports), nil
	}

	return lines, res.Reports, nil
}

func (s *RulesTestSuite) filterReports(names []string, reports []*linter.Report) []*linter.Report {
	set := make(map[string]struct{})
	for _, name := range names {
		set[name] = struct{}{}
	}

	var out []*linter.Report
	for _, r := range reports {
		if _, ok := set[r.CheckName]; ok {
			out = append(out, r)
		}
	}
	return out
}

func (s *RulesTestSuite) initEmbeddedRules(config *linter.Config) error {
	enableAllRules := func(_ rules.Rule) bool { return true }

	ruleSets, err := AddEmbeddedRules(config.Rules, enableAllRules)
	if err != nil {
		return fmt.Errorf("init embedded rules: %v", err)
	}

	for _, rset := range ruleSets {
		config.Checkers.DeclareRules(rset)
	}

	return nil
}

// createReportsByLine creates a map with a set of reports for each of the lines
// from the reports. This is necessary because there can be more than one report
// for one line.
func (s *RulesTestSuite) createReportsByLine(reports []*linter.Report) map[int][]string {
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

func (s *RulesTestSuite) ParseTestFile(l *linter.Linter, filename, content string) linter.ParseResult {
	var worker *linter.Worker
	if l.MetaInfo().IsIndexingComplete() {
		worker = l.NewLintingWorker(0)
	} else {
		worker = l.NewIndexingWorker(0)
	}

	file := workspace.FileInfo{
		Name:     filename,
		Contents: []byte(content),
	}

	var err error
	var result linter.ParseResult
	if worker.MetaInfo().IsIndexingComplete() {
		result, err = worker.ParseContents(file)
	} else {
		err = worker.IndexFile(file)
	}
	if err != nil {
		log.Fatalf("could not parse %s: %v", filename, err.Error())
	}

	return result
}
