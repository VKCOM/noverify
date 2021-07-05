// Package linttest provides linter testing utilities.
package linttest

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/VKCOM/noverify/src/workspace"
)

func init() {
	var testSeed = time.Now().UnixNano()
	if seedString := os.Getenv("TEST_SEED"); seedString != "" {
		v, err := strconv.ParseInt(seedString, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("$TEST_SEED: parse int: %v", err))
		}
		testSeed = v
	}

	rand.Seed(testSeed)
	log.Printf("TEST_SEED: %d", testSeed)
}

// SimpleNegativeTest runs linter over a single file out of given content
// and expects there to be no warnings.
//
// For positive testing, use Suite type directly.
func SimpleNegativeTest(t *testing.T, contents string) {
	t.Helper()
	if !strings.HasPrefix(contents, `<?php`) {
		t.Fatalf("PHP script doesn't start with <?php")
	}
	s := NewSuite(t)
	s.AddFile(contents)
	s.RunAndMatch()
}

// CheckFile runs linter over a single file out of given content
// and returns all reports that were found.
func CheckFile(t *testing.T, contents string) RunResult {
	s := NewSuite(t)
	s.AddFile(contents)
	return s.RunLinter()
}

// TestFile describes a file to be tested.
type TestFile struct {
	Name string

	// Data is a file contents.
	Data []byte

	// Nolint marks file as one that ignores all warnings.
	// Can be used to define builtins, for example.
	Nolint bool
}

// Suite is a configurable test runner for linter.
//
// Use NewSuite to create usable instance.
type Suite struct {
	t testing.TB

	Files  []TestFile
	Expect []string

	RuleFile string

	AllowDisable *regexp.Regexp
	MisspellList string

	LoadStubs    []string
	defaultStubs map[string]struct{}

	ignoreUndeclaredChecks bool

	config *linter.Config
	linter *linter.Linter
}

// NewSuite returns a new linter test suite for t.
func NewSuite(t testing.TB) *Suite {
	conf := linter.NewConfig()
	return &Suite{
		t: t,
		defaultStubs: map[string]struct{}{
			`stubs/phpstorm-stubs/Core/Core.php`:   {},
			`stubs/phpstorm-stubs/Core/Core_c.php`: {},
			`stubs/phpstorm-stubs/Core/Core_d.php`: {},
		},
		ignoreUndeclaredChecks: false,
		config:                 conf,
		linter:                 linter.NewLinter(conf),
	}
}

// AddFile adds a file to a suite file list.
// File gets an auto-generated name. If custom name is important,
// use AddNamedFile.
func (s *Suite) AddFile(contents string) {
	s.Files = append(s.Files, TestFile{
		Name: fmt.Sprintf("_file%d.php", len(s.Files)),
		Data: []byte(contents),
	})
}

// AddNamedFile adds a file with a specific name to a suite file list.
func (s *Suite) AddNamedFile(name, contents string) {
	s.Files = append(s.Files, TestFile{
		Name: name,
		Data: []byte(contents),
	})
}

// ReadAndAddFiles read and adds a files to a suite file list.
func (s *Suite) ReadAndAddFiles(files []string) {
	for _, f := range files {
		code, err := ioutil.ReadFile(f)
		if err != nil {
			s.t.Fatalf("read PHP file: %v", err)
		}
		s.AddNamedFile(f, string(code))
	}
}

// AddNolintFile adds a file to a suite file list that will be parsed, but not linted.
// File gets an auto-generated name. If custom name is important,
// append a properly initialized TestFile to a s Files slice directly.
func (s *Suite) AddNolintFile(contents string) {
	s.Files = append(s.Files, TestFile{
		Name:   fmt.Sprintf("_file%d.php", len(s.Files)),
		Data:   []byte(contents),
		Nolint: true,
	})
}

// RunAndMatch calls Match with the results of RunLinter.
//
// This is a recommended way to use the Suite, but if
// reports slice is needed, one can use RunLinter directly.
func (s *Suite) RunAndMatch() {
	s.t.Helper()
	s.Match(s.RunLinter().Reports)
}

// Match tries to match every report against Expect list of s.
//
// If expect slice is nil or empty, only nil (or empty) reports
// slice would match it.
func (s *Suite) Match(reports []*linter.Report) {
	expect := s.Expect
	t := s.t

	t.Helper()

	if len(reports) != len(expect) {
		t.Errorf("unexpected number of reports: expected %d, got %d",
			len(expect), len(reports))
	}

	matchedReports := map[*linter.Report]bool{}
	usedMatchers := map[int]bool{}
	for _, r := range reports {
		have := cmd.FormatReport(r)
		for i, want := range expect {
			if usedMatchers[i] {
				continue
			}
			if strings.Contains(have, want) {
				matchedReports[r] = true
				usedMatchers[i] = true
				break
			}
		}
	}
	for i, r := range reports {
		if matchedReports[r] {
			continue
		}
		t.Errorf("unexpected report %d: %s", i, cmd.FormatReport(r))
	}
	for i, want := range expect {
		if usedMatchers[i] {
			continue
		}
		t.Errorf("pattern %d matched nothing: %s", i, want)
	}

	// Only print all reports if test failed.
	if t.Failed() {
		t.Log(">>> issues reported:")
		for _, r := range reports {
			t.Log(cmd.FormatReport(r))
		}
		t.Log("<<<")
	}
}

// RunRulesTest starts testing using a file with the rules specified in RuleFile.
func (s *Suite) RunRulesTest() {
	s.t.Helper()

	if s.RuleFile == "" {
		s.t.Error("testing with rules started with an empty rule")
		return
	}

	rparser := rules.NewParser()
	rset, err := rparser.Parse("<test>", strings.NewReader(s.RuleFile))
	if err != nil {
		s.t.Fatalf("parse rules: %v", err)
	}

	s.Config().Rules = rset
	s.IgnoreUndeclaredChecks()

	ruleNamesSet := make(map[string]struct{}, len(rset.Names))
	for _, name := range rset.Names {
		ruleNamesSet[name] = struct{}{}
	}

	var filtered []*linter.Report
	result := s.RunLinter()
	for _, r := range result.Reports {
		if _, ok := ruleNamesSet[r.CheckName]; ok {
			filtered = append(filtered, r)
		}
	}

	s.Match(filtered)
}

type RunResult struct {
	Reports []*linter.Report
	Info    *meta.Info
}

// RunLinter executes linter over s Files and returns all issue reports
// that were produced during that.
func (s *Suite) RunLinter() RunResult {
	s.t.Helper()

	for _, stub := range s.LoadStubs {
		s.defaultStubs[stub] = struct{}{}
	}
	stubs := make([]string, 0, len(s.defaultStubs))
	for stub := range s.defaultStubs {
		stubs = append(stubs, stub)
	}
	if err := cmd.LoadEmbeddedStubs(s.linter, stubs); err != nil {
		s.t.Fatalf("load stubs: %v", err)
	}

	if s.MisspellList != "" {
		err := cmd.LoadMisspellDicts(s.linter.Config(), strings.Split(s.MisspellList, ","))
		if err != nil {
			s.t.Fatalf("load misspell dicts: %v", err)
		}
	}

	s.linter.UseChecks(linter.NewEnabledAllChecks())

	indexing := s.linter.NewIndexingWorker(0)
	indexing.AllowDisable = s.AllowDisable

	shuffleFiles(s.Files)
	for _, f := range s.Files {
		parseTestFile(s.t, indexing, f)
	}

	s.linter.MetaInfo().SetIndexingComplete(true)

	linting := s.linter.NewLintingWorker(0)
	linting.AllowDisable = s.AllowDisable

	shuffleFiles(s.Files)
	var reports []*linter.Report
	for _, f := range s.Files {
		if f.Nolint {
			// Mostly used to add builtin definitions
			// and for other kind of stub code that was
			// inserted to make actual testing easier (or possible, even).
			continue
		}

		result := parseTestFile(s.t, linting, f)
		reports = append(reports, result.Reports...)
	}

	declared := make(map[string]struct{})
	for _, info := range s.linter.Config().Checkers.ListDeclared() {
		declared[info.Name] = struct{}{}
	}
	if !s.ignoreUndeclaredChecks {
		for _, r := range reports {
			_, ok := declared[r.CheckName]
			if !ok {
				s.t.Errorf("got report from undeclared checker %s", r.CheckName)
			}
		}
	}

	return RunResult{
		Reports: reports,
		Info:    s.linter.MetaInfo(),
	}
}

// RunFilterLinter calls RunLinter with the filter.
func (s *Suite) RunFilterLinter(filters []string) []*linter.Report {
	s.t.Helper()
	result := s.RunLinter()

	disable := map[string]bool{}
	for _, checkName := range filters {
		disable[checkName] = true
	}
	filteredReports := result.Reports[:0]
	for _, r := range result.Reports {
		if !disable[r.CheckName] {
			filteredReports = append(filteredReports, r)
		}
	}

	return filteredReports
}

func ParseTestFile(t testing.TB, l *linter.Linter, filename, content string) linter.ParseResult {
	var worker *linter.Worker
	if l.MetaInfo().IsIndexingComplete() {
		worker = l.NewLintingWorker(0)
	} else {
		worker = l.NewIndexingWorker(0)
	}
	return parseTestFile(t, worker, TestFile{
		Name: filename,
		Data: []byte(content),
	})
}

// RunFilterMatch calls Match with the filtered results of RunLinter.
func RunFilterMatch(test *Suite, names ...string) {
	test.t.Helper()
	test.Match(filterReports(names, test.RunLinter().Reports))
}

func FindPHPFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".php") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// InitEmbeddedRules initializes embedded rules for testing.
func InitEmbeddedRules(config *linter.Config) error {
	enableAllRules := func(_ rules.Rule) bool { return true }

	ruleSets, err := cmd.AddEmbeddedRules(config.Rules, enableAllRules)
	if err != nil {
		return fmt.Errorf("init embedded rules: %v", err)
	}

	for _, rset := range ruleSets {
		config.Checkers.DeclareRules(rset)
	}

	return nil
}

func (s *Suite) UseConfig(config *linter.Config) {
	s.config = config
	s.linter = linter.NewLinter(config)
}

func (s *Suite) UseLinter(lint *linter.Linter) {
	s.linter = lint
	s.config = lint.Config()
}

func (s *Suite) Linter() *linter.Linter {
	return s.linter
}

func (s *Suite) Config() *linter.Config {
	return s.config
}

func (s *Suite) IgnoreUndeclaredChecks() {
	s.ignoreUndeclaredChecks = true
}

func filterReports(names []string, reports []*linter.Report) []*linter.Report {
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

func init() {
	var once sync.Once
	once.Do(func() { go linter.MemoryLimiterThread(0) })
}

func shuffleFiles(files []TestFile) {
	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})
}

func parseTestFile(t testing.TB, worker *linter.Worker, f TestFile) linter.ParseResult {
	file := workspace.FileInfo{
		Name:     f.Name,
		Contents: f.Data,
	}

	var err error
	var result linter.ParseResult
	if worker.MetaInfo().IsIndexingComplete() {
		result, err = worker.ParseContents(file)
	} else {
		err = worker.IndexFile(file)
	}
	if err != nil {
		t.Fatalf("could not parse %s: %v", f.Name, err.Error())
	}

	return result
}
