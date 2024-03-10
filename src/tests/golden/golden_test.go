package golden_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
)

func TestGolden(t *testing.T) {
	commandTests := []*linttest.CommandE2ETestSuite{
		{
			Name: "help",
			Args: []string{"help"},
		},
		{
			Name: "check help",
			Args: []string{"check", "help"},
		},
		{
			Name: "version",
			Args: []string{"version"},
		},
		{
			Name: "checkers",
			Args: []string{"checkers"},
		},
		{
			Name: "checkers help",
			Args: []string{"checkers", "help"},
		},
		{
			Name: "checkers name",
			Args: []string{"checkers", "arraySyntax"},
		},
	}
	targets := []*linttest.GoldenTestSuite{
		{
			Name: "embeddedrules",
			Disable: []string{
				"deadCode",
				"notStrictTypes",
				"noDeclareSection",
			},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
			},
		},

		{
			Name: "mustache",
			Disable: []string{
				"arraySyntax",
				"redundantCast",
				"notStrictTypes",
				"noDeclareSection",
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
			Disable: []string{
				"notStrictTypes",
				"noDeclareSection",
			},
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
			Disable: []string{
				"notStrictTypes",
				"noDeclareSection",
			},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/gd/gd.php`,
			},
		},

		{
			Name: "ctype",
			Disable: []string{
				"notStrictTypes",
				"noDeclareSection",
			},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
			},
		},

		{
			Name: "idn",
			Disable: []string{
				"notStrictTypes",
				"noDeclareSection",
			},
			Deps: []string{
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			Name: "parsedown",
			Disable: []string{"missingPhpdoc",
				"arraySyntax",
				"notStrictTypes",
				"noDeclareSection",
			},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			Name: "underscore",
			Disable: []string{
				"missingPhpdoc",
				"notStrictTypes",
				"noDeclareSection",
			},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
			},
		},

		{
			Name: "phprocksyd",
			Disable: []string{
				"missingPhpdoc",
				"notStrictTypes",
				"noDeclareSection",
			},
			Deps: []string{
				`stubs/phpstorm-stubs/standard/basic.php`,
				`stubs/phpstorm-stubs/pcntl/pcntl.php`,
				`stubs/phpstorm-stubs/json/json.php`,
				`stubs/phpstorm-stubs/posix/posix.php`,
			},
		},

		{
			Name: "flysystem",
			Disable: []string{
				"redundantCast",
				"notStrictTypes",
				"noDeclareSection",
			},
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
			Name: "inflector",
			Disable: []string{
				"missingPhpdoc",
				"notStrictTypes",
				"noDeclareSection",
			},
			Deps: []string{
				`stubs/phpstorm-stubs/pcre/pcre.php`,
				`stubs/phpstorm-stubs/SPL/SPL.php`,
				`stubs/phpstorm-stubs/mbstring/mbstring.php`,
			},
		},

		{
			Name: "options-resolver",
			Disable: []string{
				"missingPhpdoc",
				"notStrictTypes",
				"noDeclareSection",
			},
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
			Name: "twitter-api-php",
			Disable: []string{
				"missingPhpdoc",
				"arraySyntax",
				"notStrictTypes",
				"noDeclareSection",
			},
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
			Disable: []string{
				"notStrictTypes",
				"noDeclareSection",
			},
		},

		{
			Name: "gitignore-test",
			Disable: []string{
				"notStrictTypes",
				"noDeclareSection",
			},
			OnlyE2E:   true,
			Gitignore: true,
		},

		{
			Name: "baseline-test",
			Disable: []string{
				"notStrictTypes",
				"noDeclareSection",
			},
			OnlyE2E:  true,
			Baseline: true,
		},

		{
			Name: "phpdoc",
			Disable: []string{
				"notStrictTypes",
				"noDeclareSection",
			},
		},
	}

	e2eSuite := linttest.NewGoldenE2ETestSuite(t)

	linterConfig := linter.NewConfig("8.1")
	linterConfig.StrictMixed = true

	err := linttest.InitEmbeddedRules(linterConfig)
	if err != nil {
		t.Fatal(err)
	}

	for _, target := range targets {
		l := linter.NewLinter(linterConfig)
		linttest.PrepareGoldenTestSuite(target, t, l, "testdata", "golden.txt")
		target.Run()
		e2eSuite.RegisterTest(target)
	}
	for _, test := range commandTests {
		e2eSuite.RegisterCommandTest(test)
	}

	// Second pass only happens if none of the tests above have failed.
	if !t.Failed() {
		e2eSuite.Run()
	}
}
