package linttest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/utils"
)

// whenever dump expected output to passed golden file
const dumpExpected = false

type GoldenTestSuite struct {
	suite *Suite

	Name    string
	Deps    []string
	Disable []string

	OnlyE2E   bool
	Gitignore bool
	Baseline  bool

	SrcDir         string
	BaseDir        string
	GoldenFileName string

	// dump expected output to passed golden file
	dump bool

	// want is a golden file contents.
	want       []byte
	path       string
	reportFile *linterOutput

	// flag indicating that the structure is ready for use.
	// If the structure was created directly and the PrepareGoldenTestSuite
	// function was not called, prepared will be false, leading to panic.
	prepared bool
}

var defaultStubs = []string{
	`stubs/phpstorm-stubs/Core/Core.php`,
	`stubs/phpstorm-stubs/Core/Core_d.php`,
	`stubs/phpstorm-stubs/Core/Core_c.php`,
	`stubs/phpstorm-stubs/standard/standard_defines.php`,
	`stubs/phpstorm-stubs/standard/standard_0.php`,
	`stubs/phpstorm-stubs/standard/standard_1.php`,
	`stubs/phpstorm-stubs/standard/standard_2.php`,
	`stubs/phpstorm-stubs/standard/standard_3.php`,
	`stubs/phpstorm-stubs/standard/standard_4.php`,
	`stubs/phpstorm-stubs/standard/standard_5.php`,
	`stubs/phpstorm-stubs/standard/standard_6.php`,
	`stubs/phpstorm-stubs/standard/standard_7.php`,
	`stubs/phpstorm-stubs/standard/standard_8.php`,
	`stubs/phpstorm-stubs/standard/standard_9.php`,
}

// NewGoldenTestSuite returns a new golden test suite for t.
func NewGoldenTestSuite(t *testing.T, name, baseDir, goldenFileName string) *GoldenTestSuite {
	return &GoldenTestSuite{
		suite:          NewSuite(t),
		Name:           name,
		BaseDir:        baseDir,
		GoldenFileName: goldenFileName,
		Deps:           defaultStubs,
		prepared:       true,
	}
}

// PrepareGoldenTestSuite configures fields and standard stubs.
//
// Used if the structure was created directly.
func PrepareGoldenTestSuite(s *GoldenTestSuite, t *testing.T, l *linter.Linter, baseDir, goldenFileName string) {
	s.suite = NewSuite(t)
	if l != nil {
		s.suite.UseLinter(l)
	}
	s.BaseDir = baseDir
	s.prepared = true
	s.GoldenFileName = goldenFileName
	s.Deps = append(s.Deps, defaultStubs...)

	s.dump = dumpExpected
}

func (s *GoldenTestSuite) AddDeps(deps []string) {
	s.Deps = append(s.Deps, deps...)
}

func (s *GoldenTestSuite) AddDisabled(disabled []string) {
	s.Disable = append(s.Disable, disabled...)
}

func (s *GoldenTestSuite) Run() {
	if !s.prepared {
		panic("Structure was created directly, but the PrepareGoldenTestSuite function was not called")
	}

	s.loadGoldenFile()

	if s.OnlyE2E {
		return
	}

	runGoldenTest(s)
}

func runGoldenTest(s *GoldenTestSuite) {
	const misspellList = "Eng"

	s.suite.t.(*testing.T).Run(s.Name, func(t *testing.T) {
		phpFiles, err := utils.FindPHPFiles(s.SrcDir)
		if err != nil {
			t.Fatalf("list files: %v", err)
		}

		s.suite.LoadStubs = s.Deps
		s.suite.MisspellList = misspellList
		s.suite.ReadAndAddFiles(phpFiles)

		reports := s.suite.RunFilterLinter(s.Disable)

		s.checkGoldenOutput(s.want, reports, s.path, s.dump)
	})
}

func (s *GoldenTestSuite) loadGoldenFile() {
	path := filepath.Join(s.BaseDir, s.Name, s.GoldenFileName)
	want, err := ioutil.ReadFile(path)
	if err != nil {
		s.suite.t.Fatalf("read golden file: %v", err)
	}
	s.path = path
	s.want = want
	if s.SrcDir == "" {
		s.SrcDir = filepath.Join("testdata", s.Name)
	}
}

type linterOutput struct {
	Reports []*linter.Report
	Errors  []string
}

func (s *GoldenTestSuite) loadReportsFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		s.suite.t.Fatalf("read reports file: %v", err)
	}
	var output linterOutput
	if err := json.Unmarshal(data, &output); err != nil {
		s.suite.t.Fatalf("unmarshal reports file: %v", err)
	}
	s.reportFile = &output
}

func (s *GoldenTestSuite) checkGoldenOutput(want []byte, reports []*linter.Report, path string, dump bool) {
	haveLines := s.formatReportLines(reports)
	if dump {
		var buf bytes.Buffer
		for _, line := range haveLines {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		ioutil.WriteFile(path, bytes.TrimSuffix(buf.Bytes(), []byte("\n")), 0644)
		return
	}

	wantString := string(want)
	wantLines := strings.Split(strings.ReplaceAll(wantString, "\r", ""), "\n")

	if diff := cmp.Diff(wantLines, haveLines); diff != "" {
		s.suite.t.Errorf("results mismatch (+ have) (- want): %s", diff)
		// Use fmt.Printf() instead of t.Logf() to make the output
		// more copy/paste friendly.
		fmt.Printf("have:\n%s", strings.Join(haveLines, "\n"))
		fmt.Printf("want:\n%s", want)
	}
}

func (s *GoldenTestSuite) formatReportLines(reports []*linter.Report) []string {
	sort.SliceStable(reports, func(i, j int) bool {
		return reports[i].Filename < reports[j].Filename
	})
	var parts []string
	for _, r := range reports {
		part := strings.ReplaceAll(cmd.FormatReport(r), "\r", "")
		parts = append(parts, strings.Split(part, "\n")...)
	}
	parts = append(parts, "") // Trailing EOL
	return parts
}

type CommandE2ETestSuite struct {
	parent *GoldenE2ETestSuite

	Name string
	Args []string
}

func (s *CommandE2ETestSuite) run(t *testing.T) {
	t.Run(strings.Join(s.Args, " "), func(t *testing.T) {
		s.execHelpCommand(t, s.Args)
	})
}

func (s *CommandE2ETestSuite) execHelpCommand(t *testing.T, args []string) {
	data, err := exec.Command(s.parent.execPath, args...).CombinedOutput()
	if err != nil {
		t.Errorf("Error running command '%s': %v\nOutput:\n%s", strings.Join(args, " "), err, string(data))
	}
}

type GoldenE2ETestSuite struct {
	t *testing.T

	execPath string

	tests        []*GoldenTestSuite
	commandTests []*CommandE2ETestSuite
}

func NewGoldenE2ETestSuite(t *testing.T) *GoldenE2ETestSuite {
	return &GoldenE2ETestSuite{
		t:        t,
		execPath: "./phplinter.exe",
	}
}

func (s *GoldenE2ETestSuite) RegisterTest(test *GoldenTestSuite) {
	s.tests = append(s.tests, test)
}

func (s *GoldenE2ETestSuite) Run() {
	if testing.Short() {
		s.t.Logf("e2e is skipped in -short mode")
		return
	}

	s.BuildNoVerify()
	s.RunOnlyTests()
	s.RunCommandsTest()
	s.RemoveNoVerify()
	s.RemoveTestsFiles()
}

func (s *GoldenE2ETestSuite) BuildNoVerify() {
	goArgs := []string{
		"build",
		"-o", s.execPath,
		"-race",
		"../../../", // Using relative target to avoid problems with modules/vendor/GOPATH
	}
	out, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		s.t.Fatalf("build noverify: %v: %s", err, out)
	}
}

func (s *GoldenE2ETestSuite) RemoveNoVerify() {
	_ = os.Remove(s.execPath)
}

func (s *GoldenE2ETestSuite) RemoveTestsFiles() {
	toRemove, err := filepath.Glob("phplinter-output-*.json")
	if err != nil {
		log.Fatalf("glob: %v", err)
	}
	for _, filename := range toRemove {
		err := os.Remove(filename)
		if err != nil {
			log.Printf("tests cleanup: remove %s: %v", filename, err)
		}
	}
}

func (s *GoldenE2ETestSuite) RunOnlyTests() {
	wd, err := os.Getwd()
	if err != nil {
		s.t.Fatalf("getwd: %v", err)
	}
	wd = strings.ReplaceAll(wd, "\\", "/")

	s.t.Run("e2e", func(t *testing.T) {
		for _, test := range s.tests {
			test := test // To avoid the invalid capture in parallel tests
			t.Run(test.Name+"/e2e", func(t *testing.T) {
				t.Parallel()

				outputFilename := fmt.Sprintf("phplinter-output-%s.json", test.Name)
				args := []string{
					"check",
					"--critical", "",
					"--output-json",
					"--disable-cache", // TODO: test with cache as well
					"--allow-all-checks",
					"--strict-mixed",
					"--output", outputFilename,
				}
				if len(test.Disable) != 0 {
					args = append(args, "--exclude-checks", strings.Join(test.Disable, ","))
				}
				if test.Gitignore {
					args = append(args, "--gitignore")
				}
				if test.Baseline {
					args = append(args, "--baseline", filepath.Join("testdata", test.Name, "baseline.json"))
				}
				args = append(args, test.SrcDir)

				// Use GORACE=history_size to increase the stacktrace limit.
				// See https://github.com/golang/go/issues/10661
				phplinterCmd := exec.Command(s.execPath, args...)
				phplinterCmd.Env = append([]string{}, os.Environ()...)
				phplinterCmd.Env = append(phplinterCmd.Env, "GORACE=history_size=7")
				if len(test.Deps) != 0 {
					deps := strings.Join(test.Deps, ",")
					phplinterCmd.Env = append(phplinterCmd.Env, "NOVERIFYDEBUG_LOAD_STUBS="+deps)
				}

				out, err := phplinterCmd.CombinedOutput()
				if err != nil {
					t.Fatalf("%v: %s", err, out)
				}

				test.loadReportsFile(outputFilename)

				for _, r := range test.reportFile.Reports {
					// Turn absolute paths to something that is compatible
					// with what we get from the testdata-loaded inputs.
					r.Filename = strings.TrimPrefix(r.Filename, wd)
					// TODO: make paths absolute in non-e2e tests so we can
					// remove this "/" prefix trimming.
					r.Filename = strings.TrimPrefix(r.Filename, "/")
				}

				test.checkGoldenOutput(test.want, test.reportFile.Reports, test.path, test.dump)
			})
		}
	})
}

func (s *GoldenE2ETestSuite) RegisterCommandTest(suite *CommandE2ETestSuite) {
	suite.parent = s
	s.commandTests = append(s.commandTests, suite)
}

func (s *GoldenE2ETestSuite) RunCommandsTest() {
	s.t.Run("commands", func(t *testing.T) {
		for _, test := range s.commandTests {
			test.run(t)
		}
	})
}
