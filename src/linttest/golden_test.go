package linttest_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/rules"
	"github.com/google/go-cmp/cmp"
)

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

	targets := []struct {
		name    string
		deps    []string
		disable []string
	}{
		{
			name: "embeddedrules",
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
	}

	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			phpFiles, err := findPHPFiles(filepath.Join("testdata", target.name))
			if err != nil {
				t.Fatalf("list files: %v", err)
			}

			goldenFile := filepath.Join("testdata/", target.name, "golden.txt")
			want, err := ioutil.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("read golden file: %v", err)
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

			var parts []string
			reports := test.RunLinter()
			sort.SliceStable(reports, func(i, j int) bool {
				return reports[i].GetFilename() < reports[j].GetFilename()
			})
			for _, r := range reports {
				if disable[r.CheckName()] {
					continue
				}
				parts = append(parts, strings.Split(r.String(), "\n")...)
			}
			parts = append(parts, "") // Trailing EOL

			haveLines := parts
			wantLines := strings.Split(string(want), "\n")
			if diff := cmp.Diff(wantLines, haveLines); diff != "" {
				t.Errorf("results mismatch (+ have) (- want): %s", diff)
				// Use fmt.Printf() instead of t.Logf() to make the output
				// more copy/paste friendly.
				fmt.Printf("have:\n%s", strings.Join(haveLines, "\n"))
				fmt.Printf("want:\n%s", want)
			}
		})
	}
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
