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

/**
 * @warning 3rd argument of in_array must be true when comparing strings
 * @type string $needle
 */
in_array($needle, $_);

/**
 * @warning strings must be compared using '===' operator
 * @type string $x
 * @or
 * @type string $y
 */
$x == $y;

/**
 * @maybe did you meant to compare an object with null?
 * @type object $x
 */
$x === false;
`

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function stripos($haystack, $needle, $offset = 0) { return 0; }
function explode($delimeter, $s, $limit = 0) { return []; }
function in_array($needle, $haystack, $strict = false) { return true; }
function define($name, $value) {}

define('true', 1 == 1);
define('false', 1 == 0);

/**
 * @param Foo[] $arr
 */
function objectCompare(object $o1, Foo $o2, $x, $arr) {
  $o3 = $o1;
  $_ = $o1 === false;
  $_ = $o2 === false;
  $_ = $o3 === false;

  $o4 = $o1;
  if ($x) {
    $o4 = false;
  }
  $_ = $o4 === false;

  $int = 10;
  $_ = $int === false;
  $_ = $x === false;
  $_ = $arr === false;
}

function f($x, $y) {
  $_ = stripos("needle", $x); // Bad
  $_ = stripos($x, "needle"); // Good
  $_ = stripos($x, $y);       // Good

  $_ = $x && $x; // Bad
  $_ = 1 && $x;  // Good
  $_ = $x && $y; // Good

  $str = 'x';
  $int = 1;
  $_ = in_array('x', $x);    // Bad
  $_ = in_array($str, $x);   // Bad
  $_ = in_array('x', $x, 1); // Good
  $_ = in_array($int, $x);   // Good

  $_ = $str == '1';  // Bad
  $_ = '1' == $str;  // Bad
  $_ = $str == $x;   // Bad
  $_ = $str === '1'; // Good
}

$s = "123";
$_ = explode("", $s);
`)

	test.Expect = []string{
		`duplicated sub-expressions inside boolean expression`,
		`suspicious order of stripos function arguments`,
		`don't call explode with empty delimiter`,
		`3rd argument of in_array must be true when comparing strings`,
		`3rd argument of in_array must be true when comparing strings`,
		`strings must be compared using '===' operator`,
		`strings must be compared using '===' operator`,
		`strings must be compared using '===' operator`,
		`did you meant to compare an object with null?`,
		`did you meant to compare an object with null?`,
		`did you meant to compare an object with null?`,
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
