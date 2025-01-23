package rules_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestRuleBlock(t *testing.T) {
	rfile := `<?php
function blockEndsWithExit() {
  /** @warning block ends with exit */
  { ${"*"}; exit($_); }
}
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	test.AddFile(`<?php
{
  exit(0); // 1
}

{
  echo 123;
  exit(1); // 2
}

{}

{
  echo 1;
  exit(2);
  echo 2;
}
`)
	test.Expect = []string{
		`block ends with exit`,
		`block ends with exit`,
	}

	test.RunRulesTest()
}

func TestRuleIfElseif(t *testing.T) {
	rfile := `<?php
function testrule() {
  /** @warning bad function called */
  bad();
}
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile

	test.AddFile(`<?php
function bad() { return 0; }
function good() { return 1; }
`)

	test.AddNamedFile(`/elseif_cond1.php`, `<?php
if (good()) {
} elseif (bad()) {}`)

	test.AddNamedFile(`/elseif_cond2.php`, `<?php
if (good()) {
} elseif (bad()) {
} elseif (bad()) {}`)

	test.AddNamedFile(`/if_cond.php`, `<?php
	if (bad()) {
	} elseif (good()) {}`)

	test.Expect = []string{
		`bad function called at /elseif_cond1.php`,
		`bad function called at /elseif_cond2`,
		`bad function called at /elseif_cond2`,
		`bad function called at /if_cond.php`,
	}
	test.RunRulesTest()
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
	test.RuleFile = rfile
	code := `<?php
          $hello = 'echo 123;';
          eval($hello);
          eval('echo 456;');
        `
	test.AddNamedFile("/home/john/my/site/foo.php", code)
	test.AddNamedFile("/home/john/my/site/ads_foo.php", code)
	test.AddNamedFile("/home/john/my/site/ads_bar.php", code)

	test.Expect = []string{
		`don't eval from variable`,
		`don't eval from variable`,
	}
	test.RunRulesTest()
}

func TestRuleMultiPathFilter(t *testing.T) {
	rfile := `<?php
/**
 * @name varEval
 * @warning don't eval from variable
 * @path my/site/ads_
 * @path my/site/admin_
 */
eval(${"var"});
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	code := `<?php
          $hello = 'echo 123;';
          eval($hello);
          eval('echo 456;');
        `
	test.AddNamedFile("/home/john/my/site/foo.php", code)
	test.AddNamedFile("/home/john/my/site/ads_foo.php", code)
	test.AddNamedFile("/home/john/my/site/ads_bar.php", code)
	test.AddNamedFile("/home/john/my/site/admin_table.php", code)

	test.Expect = []string{
		`don't eval from variable`,
		`don't eval from variable`,
		`don't eval from variable`,
	}
	test.RunRulesTest()
}

func TestRulePathGroup(t *testing.T) {
	rfile := `<?php

/**
 * @path-group-name test
 * @path my/site/ads_
 */
_init_path_group_();

/**
 * @name varEval
 * @warning don't eval from variable
 * @path-group test
 * @path my/site/admin_
 */
eval(${"var"});
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	code := `<?php
          $hello = 'echo 123;';
          eval($hello);
          eval('echo 456;');
        `
	test.AddNamedFile("/home/john/my/site/foo.php", code)
	test.AddNamedFile("/home/john/my/site/ads_foo.php", code)
	test.AddNamedFile("/home/john/my/site/ads_bar.php", code)
	test.AddNamedFile("/home/john/my/site/admin_table.php", code)

	test.Expect = []string{
		`don't eval from variable`,
		`don't eval from variable`,
		`don't eval from variable`,
	}
	test.RunRulesTest()
}

func TestRulePathGroupExclude(t *testing.T) {
	rfile := `<?php
/**
 * @path-group-name test
 * @path www/no
 */
_init_path_group_();


/**
 * @name varEval
 * @warning don't eval from variable
 * @path www/
 * @path-group-exclude test
 */
eval(${"var"});
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	code := `<?php
          $hello = 'echo 123;';
          eval($hello);
          eval('echo 456;');
        `
	test.AddNamedFile("www/no", code)

	test.RunRulesTest()
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
 * @name falseCmp
 * @maybe did you meant to compare an object with null?
 * @type object $x
 */
$x === false;
`

	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	test.AddFile(`<?php
function stripos($haystack, $needle, $offset = 0) { return 0; }
function explode($delimeter, $s, $limit = 0) { return []; }
function in_array($needle, $haystack, $strict = false) { return true; }
function array_key_exists($needle, $haystack) { return false; }
function implode($glue, $pieces) { return ''; }

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
	test.RunRulesTest()
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
	test.RuleFile = rfile
	test.AddFile(`<?php
if (123); // No warning

function f() {
  if (123); // Warning
}
`)

	test.Expect = []string{
		`suspicious empty body of the if statement`,
	}
	test.RunRulesTest()
}

func TestLinkTag(t *testing.T) {
	rfile := `<?php
/**
 * @name emptyIf
 * @warning suspicious empty body of the if statement
 * @scope local
 * @link goodrule.com
 */
if ($_);
`

	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	test.AddFile(`<?php
if (123); // No warning

function f() {
  if (123); // Warning
}
`)

	test.Expect = []string{
		` | More about this rule: goodrule.com`,
	}
	test.RunRulesTest()
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
`

	test := linttest.NewSuite(t)
	test.RuleFile = rfile
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
	}
	test.RunRulesTest()
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
	test.RuleFile = rfile
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
	test.RunRulesTest()
}

func TestRulesFilter(t *testing.T) {
	rfile := `<?php
function basic_rules() {
  /**
   * @warning Don't use $var variable
   * @filter $var ^book_id$
   */
  $var;
}
function id_check() {
  /**
   * @warning Don't use the name $id for the variable
   * @filter $id ^(user|owner)_id$
   */
  any: {
    $id ==  0;
    $id === 0;
  }
}
function type_type_rules() {
  /**
   * @warning Don't use $animal variable
   * @type int $animal
   * @filter $animal ^animal_(name|id)$
   */
  $animal == 0;
}
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	test.AddFile(`<?php
function basic_test() {
  $book_id = 100;
  echo $book_id;
  echo $book_id == 0;
  if ($book_id == 0) {}
}
function test(int $user_id, int $id, int $owner_id, int $chat_id, int $id_owner) {
  $_ = $user_id === 0;
  $_ = $user_id == 0;

  $_ = $id === 0;
  $_ = $id == 0;

  $_ = $owner_id === 0;
  $_ = $owner_id == 0;

  $_ = $chat_id === 0;
  $_ = $chat_id == 0;

  $_ = $id_owner === 0;
  $_ = $id_owner == 0;
}
function type_type_check(string $animal_name, int $animal_id) {
  $_ = $animal_name == 0;

  $_ = $animal_id == 0;
}
`)
	test.Expect = []string{
		`Don't use $book_id variable`,
		`Don't use $book_id variable`,
		`Don't use $book_id variable`,
		`Don't use the name $user_id for the variable`,
		`Don't use the name $user_id for the variable`,
		`Don't use the name $owner_id for the variable`,
		`Don't use the name $owner_id for the variable`,
		`Don't use $animal_id variable`,
	}
	test.RunRulesTest()
}

func TestRulePathExcludePositive(t *testing.T) {
	rfile := `<?php
/**
 * @name varEval
 * @warning don't eval from variable
 * @path www/
 * @path-exclude www/no
 */
eval(${"var"});
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	code := `<?php
          $hello = 'echo 123;';
          eval($hello);
          eval('echo 456;');
        `
	test.AddNamedFile("www/no", code)

	test.RunRulesTest()
}

func TestRulePathExcludeNegative(t *testing.T) {
	rfile := `<?php
/**
 * @name varEval
 * @warning don't eval from variable
 * @path www/
 * @path-exclude www/no
 */
eval(${"var"});
`
	test := linttest.NewSuite(t)
	test.RuleFile = rfile
	code := `<?php
          $hello = 'echo 123;';
          eval($hello);
          eval('echo 456;');
        `
	test.AddNamedFile("www/no", code)
	test.AddNamedFile("www/yes", code)

	test.Expect = []string{
		`don't eval from variable`,
	}
	test.RunRulesTest()
}
