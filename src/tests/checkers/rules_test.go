package linttest_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/rules"
)

func TestRuleIfElseif(t *testing.T) {
	rfile := `<?php
function testrule() {
  /** @warning bad function called */
  bad();
}
`
	test := linttest.NewSuite(t)

	test.AddFile(`<?php
function bad() { return 0; }
function good() { return 1; }
`)

	linttest.AddNamedFile(test, `/elseif_cond1.php`, `<?php
if (good()) {
} elseif (bad()) {}`)

	linttest.AddNamedFile(test, `/elseif_cond2.php`, `<?php
if (good()) {
} elseif (bad()) {
} elseif (bad()) {}`)

	linttest.AddNamedFile(test, `/if_cond.php`, `<?php
	if (bad()) {
	} elseif (good()) {}`)

	test.Expect = []string{
		`bad function called at /elseif_cond1.php`,
		`bad function called at /elseif_cond2`,
		`bad function called at /elseif_cond2`,
		`bad function called at /if_cond.php`,
	}
	runRulesTest(t, test, rfile)
}

func TestRulePathFilter(t *testing.T) {
	rfile := `<?php
/**
 * @name varEval
 * @warning don't eval from variable
 * @path my/site/ads_
 */
eval(${"var"});
`
	test := linttest.NewSuite(t)
	code := `<?php
          $hello = 'echo 123;';
          eval($hello);
          eval('echo 456;');
        `
	linttest.AddNamedFile(test, "/home/john/my/site/foo.php", code)
	linttest.AddNamedFile(test, "/home/john/my/site/ads_foo.php", code)
	linttest.AddNamedFile(test, "/home/john/my/site/ads_bar.php", code)
	test.Expect = []string{
		`don't eval from variable`,
		`don't eval from variable`,
	}
	runRulesTest(t, test, rfile)
}

func TestAnyRules(t *testing.T) {
	rfile := `<?php
/**
 * @name badCond
 * @warning string value used in if condition
 * @type string $cond
 */
if ($cond) $_;

/**
 * @name typecheckOp
 * @warning increment of a non-numeric type
 * @type !(int|float) $x
 */
$x++;

function argsOrder() {
  /**
   * @warning implode() first arg must be a string and second should be an array
   * @type !string $glue
   * @or
   * @type !array $pieces
   */
  implode($glue, $pieces);

  /**
   * @warning suspicious arguments passed to array_key_exists
   * @type array $key
   * @or
   * @type !array $arr
   */
  array_key_exists($key, $arr);

  /**
   * @warning suspicious order of stripos function arguments
   */
  stripos(${"str"}, ${"*"});
}

/**
 * @name dupAndArgs
 * @warning duplicated sub-expressions inside boolean expression
 */
$x && $x;

/**
 * @name badCall
 * @warning don't call explode with empty delimiter
 * @scope any
 */
explode("", ${"*"});

/**
 * @name strictCmp
 * @warning 3rd argument of in_array must be true when comparing strings
 * @type string $needle
 */
in_array($needle, $_);

/**
 * @name strictCmp
 * @warning strings must be compared using '===' operator
 * @type string $x
 * @or
 * @type string $y
 */
$x == $y;

/**
 * @name falseCmp
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
function array_key_exists($needle, $haystack) { return false; }
function implode($glue, $pieces) { return ''; }

define('true', 1 == 1);
define('false', 1 == 0);

function stringCond(string $s) {
  if ($s !== '') { // Good
    if ($s) { // Bad
    }
  }
}

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

$_ = array_key_exists('123', [1]);   // Good
$_ = array_key_exists('123', []);    // Good
$_ = array_key_exists([1], '123');   // Bad: both args have bad type
$_ = array_key_exists([1], [1]);     // Bad: $key has bad type
$_ = array_key_exists([], [1]);      // Bad: $key has bad type (empty_array)
$_ = array_key_exists('123', '123'); // Bad: $arr has bad type

$i = 123;
$f = 1.53;
$a = [1];

$i++; // Good
$f++; // Good
$s++; // Bad
$a++; // Bad

$i--; // Good
$f--; // Good
$s--; // Bad
$a--; // Bad

$_ = implode("", []); // GOOD
$_ = implode($s, $a); // GOOD
$_ = implode($s, []); // GOOD
$_ = implode("", $a); // GOOD
$_ = implode($a, $s); // BAD:x array, string
$_ = implode($a, $a); // BAD: array, array
$_ = implode($s, $s); // BAD: string, string
$_ = implode($s, $i); // BAD: string, int
`)

	test.Expect = []string{
		`string value used in if condition`,
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
		`suspicious arguments passed to array_key_exists`,
		`suspicious arguments passed to array_key_exists`,
		`suspicious arguments passed to array_key_exists`,
		`suspicious arguments passed to array_key_exists`,
		`increment of a non-numeric type`,
		`increment of a non-numeric type`,
		`implode() first arg must be a string and second should be an array`,
		`implode() first arg must be a string and second should be an array`,
		`implode() first arg must be a string and second should be an array`,
		`implode() first arg must be a string and second should be an array`,
	}
	runRulesTest(t, test, rfile)
}

func TestLocalRules(t *testing.T) {
	rfile := `<?php
/**
 * @name emptyIf
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
 * @name selfAssign
 * @warning self-assignment
 * @scope root
 */
$x = $x;

/**
 * @name requireOnce
 * @maybe use require_once instead of require
 * @scope root
 */
require($_);

/**
 * @name dupSubExpr
 * @warning duplicated then/else parts in ternary expression
 * @scope root
 */
$_ ? $x : $x;

/**
 * @name noverifyString
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

func TestRulesIfCond(t *testing.T) {
	rfile := `<?php
/**
 * @name ifCond
 * @warning used string-typed value inside if condition
 * @type string $x
 */
if (${"x:var"}) $_;
`
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function concat($x, $y) { return $x . $y; }

function good1(string $x) {
  if ($x == '') {}
  if ($x == '');
}

function good2(string $x) {
  if ($x === '') {}
  if ($x === '');
}

/** @param float $y */
function good3(int $x, $y) {
  if ($x) {}
  if ($y) {}
  if ($x);
  if ($y);
}

function good4(array $xs) {
  global $a;
  if ($xs['a']) {}
  if ($xs) {}
  if ($a[10]) {}
}

/** @param string[] $a */
function ignored(string $x, string $y, $a) {
  if ($x || $y) {}
  if ($x . $y) {}
  if ($a[0]) {}
  if ($x .= '123') {}
  if (concat('a', 'b')) {}
}

function bad(string $x) {
  if ($x) {} // Bad 1
  if ($x);   // Bad 2
}
`)
	test.Expect = []string{
		`used string-typed value inside if condition`,
		`used string-typed value inside if condition`,
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

	ruleNamesSet := make(map[string]struct{}, len(rset.Names))
	for _, name := range rset.Names {
		ruleNamesSet[name] = struct{}{}
	}

	var filtered []*linter.Report
	for _, r := range test.RunLinter() {
		if _, ok := ruleNamesSet[r.CheckName]; ok {
			filtered = append(filtered, r)
		}
	}
	test.Match(filtered)

	linter.Rules = oldRules
}
