package linttest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/rules"
)

var PrepareForGoldenTestingHasBeenCalled int32 = 0

const PrepareForGoldenTestingNotCalledError = "Please call PrepareForGoldenTesting functions inside TestMain(m*testing.M) function to work correctly"

// PrepareForGoldenTesting prepares to run golden tests.
//
// You need to call this function in the TestMain function in the test package.
func PrepareForGoldenTesting(m *testing.M) {
	atomic.AddInt32(&PrepareForGoldenTestingHasBeenCalled, 1)
	enableAllRules := func(_ rules.Rule) bool { return true }
	p := rules.NewParser()
	linter.Rules = rules.NewSet()
	ruleSets, err := cmd.InitEmbeddedRules(p, enableAllRules)
	if err != nil {
		panic(fmt.Sprintf("init embedded rules: %v", err))
	}
	for _, rset := range ruleSets {
		linter.DeclareRules(rset)
	}

	exitCode := m.Run()

	_ = os.Remove("phplinter.exe")
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

	os.Exit(exitCode)
}

type GoldenTestSuite struct {
	t *testing.T

	Name    string
	Deps    []string
	Disable []string

	OnlyE2E   bool
	Gitignore bool
	Baseline  bool

	SrcDir  string
	BaseDir string

	// want is a golden file contents.
	want []byte

	// flag indicating that the structure is ready for use.
	// If the structure was created directly and the PrepareGoldenTestSuite
	// function was not called, prepared will be false, leading to panic.
	prepared bool
}

// NewGoldenTestSuite returns a new golden test suite for t.
func NewGoldenTestSuite(t *testing.T, name, baseDir string) *GoldenTestSuite {
	if PrepareForGoldenTestingHasBeenCalled == 0 {
		t.Fatal(PrepareForGoldenTestingNotCalledError)
	}
	return &GoldenTestSuite{
		t:       t,
		Name:    name,
		BaseDir: baseDir,
		Deps: []string{
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
		},
		prepared: true,
	}
}

// PrepareGoldenTestSuite configures fields and standard stubs.
//
// Used if the structure was created directly.
func PrepareGoldenTestSuite(s *GoldenTestSuite, t *testing.T, baseDir string) {
	if PrepareForGoldenTestingHasBeenCalled == 0 {
		t.Fatal(PrepareForGoldenTestingNotCalledError)
	}

	s.t = t
	s.BaseDir = baseDir
	s.prepared = true

	standardDeps := []string{
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

	for _, standardDep := range standardDeps {
		s.Deps = append(s.Deps, standardDep)
	}
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

func (s *GoldenTestSuite) loadGoldenFile() {
	path := filepath.Join(s.BaseDir, s.Name, "golden.txt")
	want, err := ioutil.ReadFile(path)
	if err != nil {
		s.t.Fatalf("read golden file: %v", err)
	}
	s.want = want
	if s.SrcDir == "" {
		s.SrcDir = filepath.Join("testdata", s.Name)
	}
}

func (s *GoldenTestSuite) AddDeps(deps []string) {
	for _, dep := range deps {
		s.Deps = append(s.Deps, dep)
	}
}

func (s *GoldenTestSuite) AddDisabled(disabled []string) {
	for _, disable := range disabled {
		s.Disable = append(s.Disable, disable)
	}
}

func runGoldenTest(target *GoldenTestSuite) {
	misspellList := "Eng"

	target.t.Run(target.Name, func(t *testing.T) {
		phpFiles, err := FindPHPFiles(target.SrcDir)
		if err != nil {
			t.Fatalf("list files: %v", err)
		}

		test := NewSuite(t)

		stubs := make(map[string]struct{}, len(target.Deps))
		for _, dep := range target.Deps {
			stubs[dep] = struct{}{}
		}
		test.LoadStubs = make([]string, 0, len(stubs))
		for stub := range stubs {
			test.LoadStubs = append(test.LoadStubs, stub)
		}

		for _, f := range phpFiles {
			code, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatalf("read PHP file: %v", err)
			}
			AddNamedFile(test, f, string(code))
		}
		test.MisspellList = misspellList

		disable := map[string]bool{}
		for _, checkName := range target.Disable {
			disable[checkName] = true
		}
		reports := test.RunLinter()
		filteredReports := reports[:0]
		for _, r := range reports {
			if !disable[r.CheckName] {
				filteredReports = append(filteredReports, r)
			}
		}

		checkGoldenOutput(t, target.want, filteredReports)
	})
}

type GoldenE2ETestSuite struct {
	t     *testing.T
	tests []*GoldenTestSuite
}

func NewGoldenE2ETestSuite(t *testing.T) *GoldenE2ETestSuite {
	if PrepareForGoldenTestingHasBeenCalled == 0 {
		t.Fatal(PrepareForGoldenTestingNotCalledError)
	}
	return &GoldenE2ETestSuite{
		t: t,
	}
}

func (s *GoldenE2ETestSuite) AddTest(test *GoldenTestSuite) {
	s.tests = append(s.tests, test)
}

func (s *GoldenE2ETestSuite) Run() {
	runGoldenTestsE2E(s)
}

func runGoldenTestsE2E(suite *GoldenE2ETestSuite) {
	if testing.Short() {
		suite.t.Logf("e2e is skipped in -short mode")
		return
	}

	goArgs := []string{
		"build",
		"-o", "phplinter.exe",
		"-race",
		"../../../", // Using relative target to avoid problems with modules/vendor/GOPATH
	}
	out, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		suite.t.Fatalf("build noverify: %v: %s", err, out)
	}

	wd, err := os.Getwd()
	if err != nil {
		suite.t.Fatalf("getwd: %v", err)
	}
	wd = strings.ReplaceAll(wd, "\\", "/")

	for _, target := range suite.tests {
		target := target // To avoid the invalid capture in parallel tests
		suite.t.Run(target.Name+"/e2e", func(t *testing.T) {
			t.Parallel()

			outputFilename := fmt.Sprintf("phplinter-output-%s.json", target.Name)
			args := []string{
				"--critical", "",
				"--output-json",
				"--disable-cache", // TODO: test with cache as well
				"--allow-all-checks",
				"--output", outputFilename,
			}
			if len(target.Disable) != 0 {
				args = append(args, "--exclude-checks", strings.Join(target.Disable, ","))
			}
			if target.Gitignore {
				args = append(args, "--gitignore")
			}
			if target.Baseline {
				args = append(args, "--baseline", filepath.Join("testdata", target.Name, "baseline.json"))
			}
			args = append(args, target.SrcDir)

			// Use GORACE=history_size to increase the stacktrace limit.
			// See https://github.com/golang/go/issues/10661
			phplinterCmd := exec.Command("./phplinter.exe", args...)
			phplinterCmd.Env = append([]string{}, os.Environ()...)
			phplinterCmd.Env = append(phplinterCmd.Env, "GORACE=history_size=7")
			if len(target.Deps) != 0 {
				deps := strings.Join(target.Deps, ",")
				phplinterCmd.Env = append(phplinterCmd.Env, "NOVERIFYDEBUG_LOAD_STUBS="+deps)
			}

			out, err := phplinterCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%v: %s", err, out)
			}

			output, err := readReportsFile(outputFilename)
			if err != nil {
				t.Fatalf("read output file %s: %v", outputFilename, err)
			}

			for _, r := range output.Reports {
				// Turn absolute paths to something that is compatible
				// with what we get from the testdata-loaded inputs.
				r.Filename = strings.TrimPrefix(r.Filename, wd)
				// TODO: make paths absolute in non-e2e tests so we can
				// remove this "/" prefix trimming.
				r.Filename = strings.TrimPrefix(r.Filename, "/")
			}

			checkGoldenOutput(t, target.want, output.Reports)
		})
	}
}

type linterOutput struct {
	Reports []*linter.Report
	Errors  []string
}

func readReportsFile(filename string) (*linterOutput, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var output linterOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func checkGoldenOutput(t *testing.T, want []byte, reports []*linter.Report) {
	haveLines := formatReportLines(reports)
	wantString := string(want)
	wantLines := strings.Split(strings.ReplaceAll(wantString, "\r", ""), "\n")
	if diff := cmp.Diff(wantLines, haveLines); diff != "" {
		t.Errorf("results mismatch (+ have) (- want): %s", diff)
		// Use fmt.Printf() instead of t.Logf() to make the output
		// more copy/paste friendly.
		fmt.Printf("have:\n%s", strings.Join(haveLines, "\n"))
		fmt.Printf("want:\n%s", want)
	}
}

func formatReportLines(reports []*linter.Report) []string {
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
