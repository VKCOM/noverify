package golden_test

import (
	"os"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/rules"
)

func TestMain(t *testing.M) {
	err := linttest.InitEmbeddedRules()
	if err != nil {
		panic(err)
	}

	exitCode := t.Run()

	os.Exit(exitCode)
}

func TestGolden(t *testing.T) {
	defer func(rset *rules.Set) {
		linter.Rules = rset
	}(linter.Rules)

	targets := []*linttest.GoldenTestSuite{
		{
			Name: "embeddedrules",
			Disable: []string{
				`deadCode`,
			},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
			},
		},

		{
			Name: "mustache",
			Disable: []string{
				`arraySyntax`,
				`redundantCast`,
			},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/SPL/SPL_f.php`,
				`stubs/phpstorm-stubs/json/json.php`,
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			Name: "math",
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/gmp/gmp.php`,
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/bcmath/bcmath.php`,
				`stubs/phpstorm-stubs/json/json.php`,
			},
		},

		{
			Name: "qrcode",
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/gd/gd.php`,
			},
		},

		{
			Name: "ctype",
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
			},
		},

		{
			Name: "idn",
			Deps: []string{
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			Name:    "parsedown",
			Disable: []string{`phpdoc`, `arraySyntax`},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			Name:    "underscore",
			Disable: []string{`phpdoc`},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
			},
		},

		{
			Name:    "phprocksyd",
			Disable: []string{`phpdoc`},
			Deps: []string{
				`stubs/phpstorm-stubs/standard/basic.php`,
				`stubs/phpstorm-stubs/pcntl/pcntl.php`,
				`stubs/phpstorm-stubs/json/json.php`,
				`stubs/phpstorm-stubs/posix/posix.php`,
			},
		},

		{
			Name:    "flysystem",
			Disable: []string{`redundantCast`},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
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
			Name:    "inflector",
			Disable: []string{"phpdoc"},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			Name:    "options-resolver",
			Disable: []string{"phpdoc"},
			Deps: []string{
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/Reflection/Reflection.php`,
				`stubs/phpstorm-stubs/Reflection/ReflectionClass.php`,
				`stubs/phpstorm-stubs/Reflection/ReflectionFunctionAbstract.php`,
				`stubs/phpstorm-stubs/Reflection/ReflectionFunction.php`,
				`stubs/phpstorm-stubs/Reflection/ReflectionParameter.php`,
			},
		},

		{
			Name:    "twitter-api-php",
			Disable: []string{"phpdoc", "arraySyntax"},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/date/date.php`,
				`stubs/phpstorm-stubs/hash/hash.php`,
				`stubs/phpstorm-stubs/curl/curl.php`,
				`stubs/phpstorm-stubs/curl/curl_d.php`,
			},
		},

		{
			Name: "output-test",
		},

		{
			Name:      "gitignore-test",
			OnlyE2E:   true,
			Gitignore: true,
		},

		{
			Name:     "baseline-test",
			OnlyE2E:  true,
			Baseline: true,
		},
	}

	e2eSuite := linttest.NewGoldenE2ETestSuite(t)

	for _, target := range targets {
		linttest.PrepareGoldenTestSuite(target, t, "testdata", "golden.txt")
		target.Run()
		e2eSuite.AddTest(target)
	}

	// Second pass only happens if none of the tests above have failed.
	if !t.Failed() {
		e2eSuite.Run()
	}
}
