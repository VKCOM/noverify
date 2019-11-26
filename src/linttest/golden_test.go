package linttest_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
	"github.com/google/go-cmp/cmp"
)

func TestGolden(t *testing.T) {
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
	}

	for _, target := range targets {
		t.Run(target.name, func(t *testing.T) {
			phpFiles, err := filepath.Glob("testdata/" + target.name + "/*.php")
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
			for _, r := range test.RunLinter() {
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
				t.Logf("have:\n%s", strings.Join(haveLines, "\n"))
				t.Logf("want:\n%s", want)
			}
		})
	}
}
