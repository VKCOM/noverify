package linttest_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/google/go-cmp/cmp"
)

type goldenTest struct {
	name    string
	deps    []string
	disable []string

	onlyE2E   bool
	gitignore bool
	baseline  bool

	srcDir string

	// want is a golden file contents.
	want []byte
}

func TestGolden(t *testing.T) {
	defer func(rset *rules.Set) {
		linter.Rules = rset
	}(linter.Rules)

	enableAllRules := func(_ rules.Rule) bool { return true }
	p := rules.NewParser()
	linter.Rules = rules.NewSet()
	if _, err := cmd.InitEmbeddedRules(p, enableAllRules); err != nil {
		t.Fatalf("init embedded rules: %v", err)
	}

	targets := []*goldenTest{
		{
			name: "embeddedrules",
		},

		{
			name: "mustache",
			disable: []string{
				`arraySyntax`,
				`redundantCast`,
			},
			deps: []string{
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/SPL/SPL_f.php`,
				`stubs/phpstorm-stubs/json/json.php`,
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			name: "math",
			deps: []string{
				`stubs/phpstorm-stubs/gmp/gmp.php`,
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/bcmath/bcmath.php`,
				`stubs/phpstorm-stubs/json/json.php`,
			},
		},

		{
			name: "qrcode",
			deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/gd/gd.php`,
			},
		},

		{
			name: "ctype",
			deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
			},
		},

		{
			name: "idn",
			deps: []string{
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			name:    "parsedown",
			disable: []string{`phpdoc`, `arraySyntax`},
			deps: []string{
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			name:    "underscore",
			disable: []string{`phpdoc`},
			deps:    []string{},
		},

		{
			name:    "phprocksyd",
			disable: []string{`phpdoc`},
			deps: []string{
				`stubs/phpstorm-stubs/standard/basic.php`,
				`stubs/phpstorm-stubs/pcntl/pcntl.php`,
				`stubs/phpstorm-stubs/json/json.php`,
				`stubs/phpstorm-stubs/posix/posix.php`,
			},
		},

		{
			name:    "flysystem",
			disable: []string{`redundantCast`},
			deps: []string{
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/SPL/SPL_c1.php`,
				`stubs/phpstorm-stubs/SPL/SPL_f.php`,
				`stubs/phpstorm-stubs/hash/hash.php`,
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
				`stubs/phpstorm-stubs/date/date.php`,
				`stubs/phpstorm-stubs/date/date_c.php`,
				`stubs/phpstorm-stubs/ftp/ftp.php`,
				`stubs/phpstorm-stubs/fileinfo/fileinfo.php`,
			},
		},

		{
			name:    "inflector",
			disable: []string{"phpdoc"},
			deps: []string{
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			name:    "options-resolver",
			disable: []string{"phpdoc"},
			deps: []string{
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/Reflection/Reflection.php`,
				`stubs/phpstorm-stubs/Reflection/ReflectionClass.php`,
				`stubs/phpstorm-stubs/Reflection/ReflectionFunctionAbstract.php`,
				`stubs/phpstorm-stubs/Reflection/ReflectionFunction.php`,
				`stubs/phpstorm-stubs/Reflection/ReflectionParameter.php`,
			},
		},

		{
			name:    "twitter-api-php",
			disable: []string{"phpdoc", "arraySyntax"},
			deps: []string{
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/date/date.php`,
				`stubs/phpstorm-stubs/hash/hash.php`,
				`stubs/phpstorm-stubs/curl/curl.php`,
				`stubs/phpstorm-stubs/curl/curl_d.php`,
			},
		},

		{
			name: "output-test",
		},

		{
			name:      "gitignore-test",
			onlyE2E:   true,
			gitignore: true,
		},

		{
			name:     "baseline-test",
			onlyE2E:  true,
			baseline: true,
		},
	}

	for _, target := range targets {
		goldenFile := filepath.Join("testdata/", target.name, "golden.txt")
		want, err := ioutil.ReadFile(goldenFile)
		if err != nil {
			t.Fatalf("read golden file: %v", err)
		}
		target.want = want

		if target.srcDir == "" {
			target.srcDir = filepath.Join("testdata", target.name)
		}
	}

	// Old-style tests: run tests inside the same process,
	// using the global state override.
	// This is simple and fast, makes test coverage collection
	// easier, but it can't test whether our linter can work from
	// the command-line in the same way as it does here.
	for _, target := range targets {
		if target.onlyE2E {
			continue
		}
		runGoldenTest(t, target)
	}

	// Second pass only happens if none of the tests above have failed.
	if !t.Failed() {
		runGoldenTestsE2E(t, targets)
	}
}

func runGoldenTestsE2E(t *testing.T, targets []*goldenTest) {
	if testing.Short() {
		t.Logf("e2e is skipped in -short mode")
		return
	}

	var linterName string
	if runtime.GOOS == "windows" {
		linterName = "phplinter.exe"
	} else {
		linterName = "phplinter"
	}

	goArgs := []string{
		"build",
		"-o", linterName,
		"-race",
		"../../", // Using relative target to avoid problems with modules/vendor/GOPATH
	}
	out, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("build noverify: %v: %s", err, out)
	}

	defer func() {
		_ = os.Remove("phplinter")
		_ = os.Remove("phplinter-output.json")
	}()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	wd = strings.ReplaceAll(wd, "\\", "/")

	for _, target := range targets {
		t.Run(target.name+"/e2e", func(t *testing.T) {
			args := []string{
				"--critical", "",
				"--output-json",
				"--disable-cache", // TODO: test with cache as well
				"--allow-all-checks",
				"--output", "phplinter-output.json",
			}
			if len(target.disable) != 0 {
				args = append(args, "--exclude-checks", strings.Join(target.disable, ","))
			}
			if target.gitignore {
				args = append(args, "--gitignore")
			}
			if target.baseline {
				args = append(args, "--baseline", filepath.Join("testdata", target.name, "baseline.json"))
			}
			args = append(args, target.srcDir)

			out, err := exec.Command(linterName, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("%v: %s", err, out)
			}

			output, err := readReportsFile("phplinter-output.json")
			if err != nil {
				t.Fatalf("read output file: %v", err)
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

func runGoldenTest(t *testing.T, target *goldenTest) {
	coreFiles := []string{
		`stubs/phpstorm-stubs/Core/Core.php`,
		`stubs/phpstorm-stubs/Core/Core_d.php`,
		`stubs/phpstorm-stubs/Core/Core_c.php`,
		`stubs/phpstorm-stubs/gd/gd.php`,
		`stubs/phpstorm-stubs/pcre/pcre.php`,
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

	t.Run(target.name, func(t *testing.T) {
		phpFiles, err := findPHPFiles(target.srcDir)
		if err != nil {
			t.Fatalf("list files: %v", err)
		}

		test := linttest.NewSuite(t)
		deps := target.deps
		deps = append(deps, coreFiles...)
		test.LoadStubs = deps
		for _, f := range phpFiles {
			code, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatalf("read PHP file: %v", err)
			}
			addNamedFile(test, f, string(code))
		}

		disable := map[string]bool{}
		for _, checkName := range target.disable {
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

func findPHPFiles(root string) ([]string, error) {
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
