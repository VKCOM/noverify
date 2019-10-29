package linttest_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/rules"
)

func TestAnyRules(t *testing.T) {
	rfile := `<?php
/** @warning suspicious order of stripos function arguments */
stripos(${"str"}, ${"*"});

/** @warning duplicated sub-expressions inside boolean expression */
$x && $x;

/**
 * @warning don't call explode with empty delimiter
 * @scope any
 */
explode("", ${"*"});
`

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function stripos($haystack, $needle, $offset = 0) { return 0; }
function explode($delimeter, $s, $limit = 0) { return []; }

function f($x, $y) {
  $_ = stripos("needle", $x); // Bad
  $_ = stripos($x, "needle"); // Good
  $_ = stripos($x, $y);       // Good

  $_ = $x && $x; // Bad
  $_ = 1 && $x;  // Good
  $_ = $x && $y; // Good
}

$s = "123";
$_ = explode("", $s);
`)

	test.Expect = []string{
		`duplicated sub-expressions inside boolean expression`,
		`suspicious order of stripos function arguments`,
		`don't call explode with empty delimiter`,
	}
	runRulesTest(t, test, rfile)
}

func TestLocalRules(t *testing.T) {
	rfile := `<?php
/**
 * @warning suspicious empty body of the if statement
 * @scope local
 */
if ($_);
`

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
if (123); // No warning

function f() {
  if (123); // Warning
}
`)

	test.Expect = []string{
		`suspicious empty body of the if statement`,
	}
	runRulesTest(t, test, rfile)
}

func TestRootRules(t *testing.T) {
	rfile := `<?php
/**
 * @warning self-assignment
 * @scope root
 */
$x = $x;

/**
 * @maybe use require_once instead of require
 * @scope root
 */
require($_);

/**
 * @warning duplicated then/else parts in ternary expression
 * @scope root
 */
$_ ? $x : $x;

/**
 * @info the linter is spelled NoVerify
 */
"noverify";
`

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function f1() {
  $xs = [];
  $xs[1] = $xs[1]; // No warn, since it's not inside root level

  return $xs;
}

$v = 100;
$v = $v; // Gives a warning

$_ = $v == 100 ? 'a' : 'a'; // Warning
$_ = $v == 100 ? 'a' : 'b'; // No warning

require("some_file.php");      // Warning
require_once("some_file.php"); // No warning

$name = "noverify"; // Warning
$name = "NoVerify"; // No warning
`)

	test.Expect = []string{
		`self-assignment`,
		`duplicated then/else parts in ternary expression`,
		`use require_once instead of require`,
		`the linter is spelled NoVerify`,
	}
	runRulesTest(t, test, rfile)
}

func runRulesTest(t *testing.T, test *linttest.Suite, rfile string) {
	rparser := rules.NewParser()
	rset, err := rparser.Parse("<test>", strings.NewReader(rfile))
	if err != nil {
		t.Fatalf("parse rules: %v", err)
	}
	oldRules := linter.Rules
	linter.Rules = rset
	test.RunAndMatch()
	linter.Rules = oldRules
}
