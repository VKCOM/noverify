// Package linttest provides linter testing utilities.
package linttest

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
)

// SimpleNegativeTest runs linter over a single file out of given content
// and expects there to be no warnings.
//
// For positive testing, use Suite type directly.
func SimpleNegativeTest(t *testing.T, contents string) {
	s := NewSuite(t)
	s.AddFile(contents)
	s.RunAndMatch()
}

// GetFileReports runs linter over a single file out of given content
// and returns all reports that were found.
func GetFileReports(t *testing.T, contents string) []*linter.Report {
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
	t *testing.T

	Files  []TestFile
	Expect []string

	LoadStubs []string
}

// NewSuite returns a new linter test suite for t.
func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

// AddFile adds a file to a suite file list.
// File gets an auto-generated name. If custom name is important,
// append a properly initialized TestFile to a s Files slice directly.
func (s *Suite) AddFile(contents string) {
	s.Files = append(s.Files, TestFile{
		Name: fmt.Sprintf("_file%d.php", len(s.Files)),
		Data: []byte(contents),
	})
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
	s.Match(s.RunLinter())
}

// Match tries to match every report against Expect list of s.
//
// If expect slice is nil or empty, only nil (or empty) reports
// slice would match it.
func (s *Suite) Match(reports []*linter.Report) {
	expect := s.Expect
	t := s.t

	if len(reports) != len(expect) {
		t.Errorf("unexpected number of reports: expected %d, got %d",
			len(expect), len(reports))
	}

	matchedReports := map[*linter.Report]bool{}
	usedMatchers := map[int]bool{}
	for _, r := range reports {
		have := r.String()
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
		t.Errorf("unexpected report %d: %s", i, r.String())
	}
	for i, want := range expect {
		if usedMatchers[i] {
			continue
		}
		t.Errorf("pattern %d matched nothing: %s", i, want)
	}

	// Only print all reports if test failed.
	if t.Failed() {
		t.Logf(">>> issues reported:")
		for _, r := range reports {
			t.Logf(r.String())
		}
		t.Logf("<<<")
	}
}

// RunLinter executes linter over s Files and returns all issue reports
// that were produced during that.
func (s *Suite) RunLinter() []*linter.Report {
	meta.ResetInfo()

	if len(s.LoadStubs) != 0 {
		if err := cmd.LoadEmbeddedStubs(s.LoadStubs); err != nil {
			s.t.Fatalf("load stubs: %v", err)
		}
	}
	for _, f := range s.Files {
		parseTestFile(s.t, f)
	}

	meta.SetIndexingComplete(true)

	var reports []*linter.Report
	for _, f := range s.Files {
		if f.Nolint {
			// Mostly used to add builtin definitions
			// and for other kind of stub code that was
			// inserted to make actual testing easier (or possible, even).
			continue
		}

		_, w := parseTestFile(s.t, f)
		for _, r := range w.GetReports() {
			if !r.IsDisabledByUser() {
				reports = append(reports, r)
			}
		}
	}

	return reports
}

// ParseTestFile parses given test file.
func ParseTestFile(t *testing.T, filename, content string) (rootNode node.Node, w *linter.RootWalker) {
	return parseTestFile(t, TestFile{
		Name: filename,
		Data: []byte(content),
	})
}

func init() {
	var once sync.Once
	once.Do(func() { go linter.MemoryLimiterThread() })
}

func parseTestFile(t *testing.T, f TestFile) (rootNode node.Node, w *linter.RootWalker) {
	var err error
	rootNode, w, err = linter.ParseContents(f.Name, f.Data, nil)
	if err != nil {
		t.Fatalf("could not parse %s: %v", f.Name, err.Error())
	}

	if !meta.IsIndexingComplete() {
		w.UpdateMetaInfo()
	}

	return rootNode, w
}
