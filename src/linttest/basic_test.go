package linttest_test

import (
	"log"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
)

func TestStrictCmp(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{`stubs/phpstorm-stubs/Core/Core_d.php`}
	test.AddFile(`<?php
function f($x) {
  $_ = ($x == false);
  $_ = (false == $x);
  $_ = ($x == true);
  $_ = (true == $x);
  $_ = ($x == null);
  $_ = (null == $x);
  return true;
}

$_ = (f(0) != false);
$_ = (false != f(0));
$_ = (f(0) != true);
$_ = (true != f(0));
$_ = (f(0) != null);
$_ = (null != f(0));
`)
	test.Expect = []string{
		`non-strict comparison with false (use ===)`,
		`non-strict comparison with false (use ===)`,
		`non-strict comparison with false (use !==)`,
		`non-strict comparison with false (use !==)`,
		`non-strict comparison with true (use ===)`,
		`non-strict comparison with true (use ===)`,
		`non-strict comparison with true (use !==)`,
		`non-strict comparison with true (use !==)`,
		`non-strict comparison with null (use ===)`,
		`non-strict comparison with null (use ===)`,
		`non-strict comparison with null (use !==)`,
		`non-strict comparison with null (use !==)`,
	}
	test.RunAndMatch()
}

func TestRedundantGlobal(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$foo = 0;

function f1() {
  global $GLOBALS;
  global $_GET;
  global $foo; // OK
  return $foo;
}

function f2() {
  global $_POST, $foo, $_ENV;
  return $foo;
}
`)
	// A full warning message is `redundantGlobal: $varname is superglobal`.
	test.Expect = []string{
		`GLOBALS is superglobal`,
		`_GET is superglobal`,
		`_POST is superglobal`,
		`_ENV is superglobal`,
	}
	test.RunAndMatch()
}

func TestForeachEmpty(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$xs = [];
foreach ($xs as $k => $v) {
  $_ = [$k, $v];
}
$_ = [$k, $v]; // Bad
foreach ($xs as $x) {
  $_ = [$x];
}
$_ = [$x]; // Bad
`)
	test.Expect = []string{
		`Variable might have not been defined: k`,
		`Variable might have not been defined: v`,
		`Variable might have not been defined: x`,
	}
	test.RunAndMatch()
}

func TestBareTry(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
try {
  echo 123;
}
`)
	test.Expect = []string{
		`At least one catch or finally block must be present`,
	}
	test.RunAndMatch()
}

func TestLinterDisable(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
/** @linter disable */
$_ = array(1);
`)
}

func TestKeywordCaseElseif(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`<?php
function f($cond) {
  if ($cond+0) {
  } Else  If ($cond+1) {
  } elsE/**/IF ($cond+2) {
  } elseiF ($cond+3) {
  } else /*a*/ /*b*/  iF ($cond+4) {
  } ElsE {}
}
`)
	test.Expect = []string{
		`Use if instead of If`,
		`Use if instead of IF`,
		`Use if instead of iF`,
		`Use else instead of Else`,
		`Use else instead of elsE`,
		`Use elseif instead of elseiF`,
		`Use else instead of ElsE`,
	}
	test.RunAndMatch()
}

func TestKeywordCase(t *testing.T) {
	test := linttest.NewSuite(t)

	// TODO:
	// - "as" in foreach; no clear way to get the pos range
	// - "class"; because of modifiers
	// - "function" for methods; because of modifiers
	// - "const" inside classes; because of modifiers
	// - "instanceof"; located in between 2 operands
	// - "while" in do
	// - "endif" and other "end*" tokens
	// - "insteadof" from trait adaptations
	// - "from" from yield-from

	test.AddFile(`<?php
Namespace Foo;
Include '.';
Include_Once '.';
Require '.';
Require_Once '.';
class TheBase {}
CONST  C1 = 1;
ABSTRACT Final class TheClass  extendS  TheBase {
  Const C2 = 2;
}
class NonAbstract {}
FOREACH ([] as $_) {}
whilE (0) { breaK; }
$a = NeW NonAbstract();
$b = CLONE  $a;
$b = Clone($a);
FUNCTION f() {
  GLOBAL $xx;
  While (0) { BREAK; }
  wHILE (0) { CONTINUE; }
  SWITCH ($xx) {
  Case 1: Break;
  DEFAULT: return 0;
  }
  if (0) {
  } ELSEIF (1) {
  } ELSE {}
  Do {
  } While (0);
  DO {} WHILE (0);
  switch (0):
  ENDswitch;
  Goto label;
  label:
  YIELD 'yelling!';
  yielD FROM 'blah!';
  FOR (;;) {}
  for (;;):
  EndFor;
  if (0):
  ENDIF;
  TRY {
  } CATCH (Exception $_) {
  } FINALLY {
  }
  $_ = $xx InstanceOf TheClass;
  $_ = Function () {};
  Return(1);
}
TRAit TheTrait1 {
  /***/
  public function f() {}
}
trait TheTrait2 {
  /***/
  public function f() {}
}
Interface TheInterface {
  PubliC function f();
}
class UsingTrait IMPLEMENTs TheInterface {
  Var $xdd;
  USE TheTrait1, TheTrait2 {
    TheTrait1::f Insteadof TheTrait2;
  }
}
THrow new NonAbstract();
function good() {
  switch (0):
  endswitch;
  foreach ([] as /* aS */ $_) {}
  foreach ([] as $_) {} // aS
  foreach ([] as $_) /* aS */ {}
  return(1);
}
`)

	test.Expect = []string{
		`Use abstract instead of ABSTRACT`,
		`Use var instead of Var`,
		`Use break instead of BREAK`,
		`Use break instead of Break`,
		`Use break instead of breaK`,
		`Use case instead of Case`,
		`Use catch instead of CATCH`,
		`Use clone instead of CLONE`,
		`Use clone instead of Clone`,
		`Use const instead of CONST`,
		`Use continue instead of CONTINUE`,
		`Use default instead of DEFAULT`,
		`Use do instead of DO`,
		`Use do instead of Do`,
		`Use else instead of ELSE`,
		`Use elseif instead of ELSEIF`,
		`Use extends instead of extendS`,
		`Use final instead of Final`,
		`Use finally instead of FINALLY`,
		`Use for instead of FOR`,
		`Use foreach instead of FOREACH`,
		`Use function instead of FUNCTION`,
		`Use global instead of GLOBAL`,
		`Use goto instead of Goto`,
		`Use implements instead of IMPLEMENTs`,
		`Use include instead of Include`,
		`Use include_once instead of Include_Once`,
		`Use interface instead of Interface`,
		`Use namespace instead of Namespace`,
		`Use new instead of NeW`,
		`Use require instead of Require`,
		`Use require_once instead of Require_Once`,
		`Use return instead of Return`,
		`Use throw instead of THrow`,
		`Use trait instead of TRAit`,
		`Use try instead of TRY`,
		`Use use instead of USE`,
		`Use while instead of While`,
		`Use while instead of wHILE`,
		`Use while instead of whilE`,
		`Use yield instead of YIELD`,
		`Use yield instead of yielD`,
		`Use public instead of PubliC`,
	}

	test.RunAndMatch()
}

func TestCallStaticParent(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class Base { protected function f() { return 1; } }
	class Derived extends Base {
		private function g() {
			return parent::f() + 1;
		}
	}
`)
	runFilterMatch(test, "callStatic")
}

func TestVoidResultUsedInAssignment(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	/**
	* @return void
	*/
	function f() {}
	$_ = f();
`)
	test.Expect = []string{
		`void function result used`,
	}
	test.RunAndMatch()
}

func TestVoidResultUsedInBinary(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function define($_, $_) {}
	define('false', 1 == 0);
	define('true', 1 != 0);

	/**
	 * @return void
	 */
	function f() {}

	$_ = f() % 2;
	$_ = f() & 1;
	$_ = f() | 1;
	$_ = f() ^ 1;
	$_ = f() && true;
	$_ = f() || true;
	$_ = (f() xor true);
	$_ = f() + 1;
	$_ = f() - 1;
	$_ = f() * 1;
	$_ = f() / 1;
	$_ = f() % 2;
	$_ = f() ** 2;
	$_ = f() == 1;
	$_ = f() != 1;
	$_ = f() === 1;
	$_ = f() !== 1;
	$_ = f() < 1;
	$_ = f() <= 1;
	$_ = f() > 1;
	$_ = f() >= 1;
`)
	test.Expect = []string{
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,
		`void function result used`,

		// $x = void() xor $y;
		//      ^^^^^^ 1st void warning
		// ^^^^^^^^^^^ 2nd void warning
		// TODO: do we want to reduce these 2 warnings into a single warning?
		`void function result used`,
		`void function result used`, // 1 extra warning is tolerated for now...
	}
	test.RunAndMatch()
}

func TestVoidParam(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	/**
	* @param void $x
	* @param int $y
	* @return void
	*/
	function f($x, $y) {}
`)
	test.Expect = []string{
		`void is not a valid type for input parameter`,
	}
	test.RunAndMatch()
}

func TestCallStatic(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class T {
		public static function sf($_) {}
		public function f($_) {}
	}
	$v = new T();
	$v->sf(1);
	T::f(1);
	`)
	test.Expect = []string{
		`Calling static method as instance method`,
		`Calling instance method as static method`,
	}
	runFilterMatch(test, "callStatic")
}

func TestForeachList(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php

foreach ([[1, 2]] as list($x, $y)) {
  $_ = [$x => $y];
}

foreach ([[1, 2, 3, 4]] as list($x, $y,,$z)) {
  $_ = [$x => $y, 5 => $z];
}
`)
}

func TestArgsCount(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function f() {
$_ = mt_rand();        // OK
$_ = mt_rand(1);       // Not OK
$_ = mt_rand(1, 2);    // OK
$_ = mt_rand(1, 2, 3); // Not OK
}

function mt_rand($x = 0, $y = 0) { return 1; }`)
	test.Expect = []string{
		`mt_rand expects 0 or 2 args`,
		`mt_rand expects 0 or 2 args`,
	}
	test.RunAndMatch()
}

func TestArgsArraysSyntax(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function bad($a = array()) {}
function good($a = []) {}
`)
	test.Expect = []string{
		`Use of old array syntax (use short form instead)`,
	}
	test.RunAndMatch()
}

func TestMethodComplexity(t *testing.T) {
	funcCode := strings.Repeat("$_ = 0;\n", 9999)
	test := linttest.NewSuite(t)
	test.AddFile(`<?php class C { private function f() {` + funcCode + `} }`)
	test.Expect = []string{"Too big method: more than 150"}
	test.RunAndMatch()
}

func TestFuncComplexity(t *testing.T) {
	funcCode := strings.Repeat("$_ = 0;\n", 9999)
	test := linttest.NewSuite(t)
	test.AddFile(`<?php function f() {` + funcCode + `}`)
	test.Expect = []string{"Too big function: more than 150"}
	test.RunAndMatch()
}

func TestPrecedenceBadLHS(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function lhs($x, $mask) {
  $_ = 0 == $mask & $x;
  $_ = 0 != $mask & $x;
  $_ = 0 === $mask & $x;
  $_ = 0 !== $mask & $x;

  $_ = 0 == $mask | $x;
  $_ = 0 != $mask | $x;
  $_ = 0 === $mask | $x;
  $_ = 0 !== $mask | $x;
}
`)
	test.Expect = []string{
		`== has higher precedence than &`,
		`!= has higher precedence than &`,
		`=== has higher precedence than &`,
		`!== has higher precedence than &`,
		`== has higher precedence than |`,
		`!= has higher precedence than |`,
		`=== has higher precedence than |`,
		`!== has higher precedence than |`,
	}
	test.RunAndMatch()
}

func TestPrecedenceBadRHS(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function rhs($x, $mask) {
  $_ = $x & $mask == 0;
  $_ = $x & $mask != 0;
  $_ = $x & $mask === 0;
  $_ = $x & $mask !== 0;

  $_ = $x | $mask == 0;
  $_ = $x | $mask != 0;
  $_ = $x | $mask === 0;
  $_ = $x | $mask !== 0;
}
`)
	test.Expect = []string{
		`== has higher precedence than &`,
		`!= has higher precedence than &`,
		`=== has higher precedence than &`,
		`!== has higher precedence than &`,
		`== has higher precedence than |`,
		`!= has higher precedence than |`,
		`=== has higher precedence than |`,
		`!== has higher precedence than |`,
	}
	test.RunAndMatch()
}

func TestPrecedenceGood(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function foo() { return 10; }

function rhs($x, $mask) {
  $_ = ($x & $mask) == 0;
  $_ = ($x & $mask) != 0;
  $_ = ($x & $mask) === 0;
  $_ = ($x & $mask) !== 0;

  $_ = ($x | $mask) == 0;
  $_ = ($x | $mask) != 0;
  $_ = ($x | $mask) === 0;
  $_ = ($x | $mask) !== 0;

  $_ = 0x02 | (($x & $mask) != 0);
  $_ = 0x02 & (foo() !== 0);
}

function lhs($x, $mask) {
  $_ = 0 == ($mask & $x);
  $_ = 0 != ($mask & $x);
  $_ = 0 === ($mask & $x);
  $_ = 0 !== ($mask & $x);

  $_ = 0 == ($mask | $x);
  $_ = 0 != ($mask | $x);
  $_ = 0 === ($mask | $x);
  $_ = 0 !== ($mask | $x);

  $_ = (($x & $mask) != 0) | 0x02;
  $_ = (foo() !== 0) & 0x02;
}
`)
}

func TestBitwiseOps(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$x = 10;
$_ = ($x > 0 & $x != 15);
$_ = ($x == 1 | $x == 2);
`)
	test.Expect = []string{
		`Used & bitwise op over bool operands, perhaps && is intended?`,
		`Used | bitwise op over bool operands, perhaps || is intended?`,
	}
	test.RunAndMatch()
}

func TestArgvGlobal(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`<?php
// OK - accessed from the global scope.
$_ = $argv[0];
$_ = $argc;

function f_good() {
  // OK - used "global" with argc and argv.
  global $argv;
  global $argc;
  $_ = $argv[0];
  $_ = $argc;
}

class Foo {
  // Same as with functions.
  public function method() {
    global $argv;
    global $argc;
    $_ = $argv[1];
    $_ = $argc;
  }
}`)
	test.AddFile(`<?php
function f_bad() {
  // Not OK - need to use "global".
  $_ = $argv[1];
}

class Foo {
  // Same as with functions.
  public function method() {
    $_ = $argc;
  }
}
`)

	test.Expect = []string{
		"Undefined variable: argv",
		"Undefined variable: argc",
	}

	runFilterMatch(test, "undefined")
}

func TestAutogenSkip(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
// auto-generated file, DO NOT EDIT!
$_ = array();`)

	linttest.SimpleNegativeTest(t, `<?php
// auto-generated file
// DO NOT EDIT!

$_ = array();`)

	linttest.SimpleNegativeTest(t, `<?php

/*
 * AUTO-GENERATED
 *
 * DO NOT EDIT UNLESS YOU KNOW WHAT YOU'RE DOING!
 */

$_ = array();`)

	linttest.SimpleNegativeTest(t, `<?php
// autogenerated (DO NOT EDIT)

$_ = array();`)

	linttest.SimpleNegativeTest(t, `<?php
/* autogenerated (DO NOT EDIT) */

$_ = array();`)

	linttest.SimpleNegativeTest(t, `<?php
// File generated by foobar.
// Do not edit (re-run generator instead).

$_ = array();`)

	linttest.SimpleNegativeTest(t, `<?php
// Do not edit this file.
// It's auto-generated.

$_ = array();`)

	linttest.SimpleNegativeTest(t, `<?php
// This file is auto-generated.
//
// This comment contains a few extra lines of text.
// Autogen files headers really shouldn't be that complex.
// Why can't we have some standard format, like in Go?
// (https://github.com/golang/go/issues/13560)
//
// DO NOT EDIT though.
$_ = array();`)

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$_ = "autogenerated; do not edit";

$_ = array();`)
	test.Expect = []string{
		`Use of old array syntax (use short form instead)`,
	}
	test.RunAndMatch()
}

func TestAssignmentsInForLoop(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function increment($i) { return $i + 1; }

for ($i = 0; $i <= 10; $i = increment($i)) {}
for ($i = increment(0); $i <= 10; $i = $i + 1) {}
for ($i = 0; $i == 0; $i++) {}
for ($i = 0; $i == 0; ++$i) {}
for ($i = 0; $i == 0; $i = $i++) {}
`)
}

func TestCustomUnusedVarRegex(t *testing.T) {
	defer func(isDiscardVar func(string) bool) {
		linter.IsDiscardVar = isDiscardVar
	}(linter.IsDiscardVar)

	linter.IsDiscardVar = func(s string) bool {
		return strings.HasPrefix(s, "_")
	}

	linttest.SimpleNegativeTest(t, `<?php
$_unused = 10;

function f() {
  $_unused2 = 20;
  $_ = 30;
  foreach ([1] as $_ => $_user) {}
}
`)
}

func TestClosureCapture(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class omg {
		public $some_property;
	}

	function doSomething($a, omg $b) {
		return function() use($b) {
			echo $b->some_property;
			echo $b->other_property;
		};
	}`)
	test.Expect = []string{"other_property does not exist"}
	test.RunAndMatch()
}

func TestOrDie1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
global $ok;
$ok or die("not ok");
echo "quite reachable\n";
`)
}

func TestOrDie2(t *testing.T) {
	// Check that we still check "or" LHS and RHS properly.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$undef1 or die($undef2);
`)
	test.Expect = []string{
		"Undefined variable: undef1",
		"Undefined variable: undef2",
	}
	test.RunAndMatch()
}

func TestOrExit(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
global $ok;
$ok or exit("");
echo "quite reachable\n";
`)
}

func TestUnusedInInstanceof(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {}

function f1($cond) {
  global $g;
  $x = $g;
  if ($x instanceof Foo) {
    // Do nothing.
  }
  if ($cond) {
    $_ = $x; // Use $x
    if ($x instanceof Foo) {
      // Should not warn about unused var.
    }
  }
}

function f2() {
  global $v;
  return $v instanceof Foo;
}

function f3() {
  global $v;
  if ($v instanceof Foo) {
    return 1;
  }
  return 0;
}
`)
}

func TestUnusedInVarPropFetch(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {}
function foo(Foo $x) {
	$y = "propname";
	return $x->$y;
}
	`)
}

func TestUnusedInVarPropAssign(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {}
function foo(Foo $x) {
	$y = "propname";
	$x->$y = "propval";
}
	`)
}

func TestStaticPropFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {}
function foo() {
	$x = "propname";
	return Foo::$$y; // $y is undefined, but $x is defined
}
`)
	test.Expect = []string{
		`Variable x is unused`,
		`Undefined variable: y`,
	}
	test.RunAndMatch()
}

func TestUnusedInStaticVarPropFetch(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {}
function foo() {
	$x = "propname";
	return Foo::$$x;
}
`)
}

func TestUnusedInStaticVarPropAssign(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {}
function foo() {
	$x = "propname";
	Foo::$$x = "propval";
}
`)
}

func TestUnusedInSwitch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function f($a) {
		switch ($a) {
		case 0:
			$x = 0; // Warning
		}
	}
	function nested($a) {
		for ($i = 0; $i < 3; $i++) {
			switch ($a) {
			case 0:
				$j = 10; // Inside loop, no warning
			}
		}
	}
	function nested2($a) {
		for ($i = 0; $i < 3; $i++) {
			switch ($a + $a) {
			case 0:
				switch ($a) {
				case 0:
					$j = 10; // Inside loop, no warning
				}
			}
		}
	}
	function insideCase($a, $b) {
		$b2 = $b;
		switch ($a) {
		case $b2['key']:
			return 10;
		}
		return 20;
	}`)
	test.Expect = []string{`Variable x is unused`}
	runFilterMatch(test, "unused")
}

func TestSwitchContinue1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	global $x;
	global $y;

	switch ($x) {
	case 10:
		continue;
	}

	switch ($x) {
	case 10:
		if ($x == $y) {
			continue;
		}
	}

	for ($i = 0; $i < 10; $i++) {
		switch ($i) {
		case 5:
			continue;
		}
	}`)
	test.Expect = []string{
		`'continue' inside switch is 'break'`,
		`'continue' inside switch is 'break'`,
		`'continue' inside switch is 'break'`,
	}
	test.RunAndMatch()
}

func TestSwitchContinue2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	global $x;
	switch ($x) {
	case 10:
		for ($i = 0; $i < 10; $i++) {
			if ($i == $x) {
				continue; // OK, bound to 'for'
			}
		}
	}

	// OK, "continue 2" does the right thing.
	// Phpstorm finds incorrect label "level" values,
	// but it doesn't report 'continue' (without level) as being bad.
	for ($i = 0; $i < 3; $i++) {
		switch ($x) {
		case 10:
			continue 2;
		}
	}`)
	test.RunAndMatch()
}

func TestBuiltinConstant(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function f() {
		$_ = NULL;
		$_ = True;
		$_ = FaLsE;
	}`)
	test.Expect = []string{
		"Use null instead of NULL",
		"Use true instead of True",
		"Use false instead of FaLsE",
	}
	test.RunAndMatch()
}

func TestFunctionNotOnlyExits2(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function rand() {
		return 4;
	}

	class RuntimeException {}

	class Something {
		/** may throw */
		public static function doExit() {
			if (rand()) {
				throw new \RuntimeException("OMG");
			}

			return rand();
		}
	}

	function doSomething() {
		Something::doExit();
		echo "Not always dead code";
	}`)
}

func TestArrayAccessForClass(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class three {}
	class five {}
	function test() {
		$a = 1==2 ? new three : new five;
		return $a['test'];
	}`)
	test.Expect = []string{"Array access to non-array type"}
	test.RunAndMatch()
}

// This test checks that expressions are evaluated in correct order.
// If order is incorrect then there would be an error that we are referencing elements of a class
// that does not implement ArrayAccess.
func TestCorrectTypes(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class three {}
	class five {}
	function test() {
		$a = ['test' => 1];
		$a = ($a['test']) ? new three : new five;
		return $a;
	}`)
}

func TestAllowReturnAfterUnreachable(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function unreachable() {
		exit;
	}

	function test() {
		unreachable();
		return;
	}`)
}

func TestFunctionReferenceParams(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function doSometing(&$result) {
		$result = 5;
	}`)
}

func TestFunctionReferenceParamsInAnonymousFunction(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function doSometing() {
		return function() use($a, &$result) {
			echo $a;
			$result = 1;
		};
	}`)
	test.Expect = []string{"Undefined variable a"}
	test.RunAndMatch()
}

func TestFunctionCallSplatArg(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function doSomething($a, $b, $c) {}
$x = [1, 2, 3];
doSomething(...$x);
	`)
}

func TestForeachByRef(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
$xs = [1, 2];
foreach ($xs as &$x) {
    if ($x) {
        $_ = $x;
    }
}
foreach ($xs as &$x) {
    $_ = $x;
}
`)
}

func TestForeachByRefUnused(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class SomeClass {
		public $a;
	}

	/**
	 * @param SomeClass[] $some_arr
	 */
	function doSometing($some_arr) {
		$some_arr = [];

		foreach ($some_arr as $var) {
			$var->a = 1;
		}

		foreach ($some_arr as &$var2) {
			$var2->a = 2;
		}
	}`)
}

func TestAllowAssignmentInForLoop(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function test() {
	  for ($day = 0; $day <= 100; $day = $day + 1) {
		echo $day;
	  }
	}
	`)
}

func TestDuplicateArrayKeyGood(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
$valid_quotes = [
  '"' => 1,
  "'" => 1,
];
`)
}

func TestDuplicateArrayKey(t *testing.T) {
		test := linttest.NewSuite(t)
		test.AddFile(`<?php
class T {
	public const C1 = 3;
	public const C2 = 4;
}

const C1 = 1;
const C2 = 2;

function id($x) {
  return $x;
}

$var = 1;
function notPure() {
  global $var;
  $var = $var + 1;
  return $var;
}

// Strings
$example5 = [
  "one" => 1,
  "two" => 2,
  "one" => 3,
];

// Integers in different base
$example1 = [
  1 => 1,
  2 => 2,
  1 => 3,
];

// Doubles
$example2 = [
  1.0 => 2,
  2.0 => 3,
  1.0 => 1,
];

// Constants
$example3 = [
  C1 => 1,
  C2 => 2,
  C1 => 3, // Duplicate key C1
];

// Class constants
$example4 = [
  T::C1 => 1,
  T::C2 => 2,
  T::C2 => 3, // Duplicate key T1::C1
];

// Should be another warning, I'll skip it explicitly
// But this is still debatable
$example6 = [
  new T() => 1,
  new T() => 2,
];

// Side effects free functions
$example7 = [
  id(1) => 1,
  id(3) => 2,
  id(1) => 3,
];

$s = "42";
// More of them
$example8 = [
  'a' . $s => 1,
  'b' . $s => 2,
  'a' . $s => 3,
];

// Not side effect free functions
$example8 = [
  notPure() => 1,
  id(1)     => 2,
  notPure() => 3,
  // another level of pureness check
  id(notPure()) => 4,
  id(notPure()) => 5,
];

$example9 = [
  $example5["one"] => 1,
  $example5["two"] => 2,
  $example5["one"] => 3,
];
`)
	test.Expect = []string{
		`Duplicate array key one`,
		`Duplicate array key 1`,
		`Duplicate array key 1.0`,
		`Duplicate array key C1`,
		`Duplicate array key T::C2`,
		`Duplicate array key id(1)`,
		`Duplicate array key 'a' . $s`,
		`Duplicate array key $example5["one"]`,
	}
	test.RunAndMatch()
}

func TestMixedArrayKeys(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function test() {
	  return [
		  'something',
		  'key2' => 'other_thing',
		  'key3' => 'third_thing',
	  ];
	}
	`)
	test.Expect = []string{"Mixing implicit and explicit array keys"}
	test.RunAndMatch()
}

func TestStringGlobalVarName(t *testing.T) {
	// Should not panic.
	linttest.SimpleNegativeTest(t, `<?php
	function f() {
		global ${"x"};
		global ${"${x}_{$x}"};
	}`)
}

func TestArrayLiteral(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function traditional_array_literal() {
		return array(1, 2);
	}`)
	test.Expect = []string{"Use of old array syntax"}
	test.RunAndMatch()
}

func TestNonEmptyVar(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function non_empty_var() {
		if (!empty($x)) {
			return $x;
		}
		return 0;
	}

	function empty_arg() {
		$_ = !empty($x) || empty($x);
	}
`)
}

func TestEmptyVar(t *testing.T) {
	// Only !empty marks a variable in a same way as isset does.

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function empty_var() {
		if (empty($x1)) {
			return $x1;
		}
		return 0;
	}
	function use_outside_if() {
		if (!empty($x2)) {
			$_ = $x2;
		}
		return $x2;
	}`)
	test.Expect = []string{
		`Undefined variable: x1`,
		`Undefined variable: x2`,
	}
	test.RunAndMatch()
}

func TestIssetElseif1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function f() {
  if (isset($x)) {
    echo $x;
  } elseif (isset($y)) {
    echo $y; // OK to use here.
  }
  echo $y; // But should be undefined here.
}
`)
	test.Expect = []string{`Undefined variable: y`}
	test.RunAndMatch()
}

func TestIssetElseif2(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
if (isset($x)) {
  echo $x;
} else if (isset($y)) {
  echo $y;
}`)
}

func TestUnused(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function unused_test($arg1, $arg2) {
		global $g;

		$_SERVER['test'] = 1; // superglobal, must not count as unused

		$_ = 'should not count as unused';
		$a = 10;
		foreach ([1, 2, 3] as $k => $v) {
			// $v is unused here
			echo $k;
		}
	}`)
	test.Expect = []string{
		"Variable g is unused",
		"Variable a is unused",
		"Variable v is unused",
	}
	test.RunAndMatch()
}

func TestAtVar(t *testing.T) {
	// variables declared using @var should not be overridden
	_ = linttest.GetFileReports(t, `<?php
	function test() {
		/** @var string $a */
		$a = true;
		return $a;
	}`)

	fi, ok := meta.Info.GetFunction(`\test`)
	if !ok {
		t.Errorf("Could not get function test")
	}

	typ := fi.Typ
	hasBool := false
	hasString := false

	typ.Iterate(func(typ string) {
		if typ == "string" {
			hasString = true
		} else if typ == "bool" {
			hasBool = true
		}
	})

	log.Printf("$a type = %s", typ)

	if !hasBool {
		t.Errorf("Type of variable a does not have boolean type")
	}

	if !hasString {
		t.Errorf("Type of variable a does not have string type")
	}
}

func TestFunctionExit(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php function doExit() {
		exit;
	}

	function doSomething() {
		doExit();
		echo "Dead code";
	}`)
	test.Expect = []string{"Unreachable code"}
	test.RunAndMatch()
}

func TestFunctionDie(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php function doDie() {
		die("123");
		echo "Also unreachable";
	}

	function doSomething() {
		doDie();
		echo "Dead code";
	}`)
	test.Expect = []string{
		"Unreachable code",
		"Unreachable code",
	}
	test.RunAndMatch()
}

func TestFunctionNotOnlyExits(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php function rand() {
		return 4;
	}

	function doExit() {
		if (rand()) {
			exit;
		} else {
			return;
		}
	}

	function doSomething() {
		doExit();
		echo "Not always dead code";
	}`)
}

func TestFunctionJustReturns(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php function justReturn() {
		return 1;
	}

	function doSomething() {
		$_ = justReturn();
		echo "Just normal code";
	}`)
}

func TestSwitchFallthrough(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function withFallthrough($a) {
		switch ($a) {
		case 1:
			echo "1\n";
			// With prepended comment line.
			// fallthrough
		case 2:
			echo "2\n";
			// falls through and continue rolling
		case 3:
			echo "3\n";
			/* fallthrough and blah-blah */
		case 4:
			echo "4\n";
			/* falls through */
		default:
			echo "Other\n";
		}
	}`)
}

func TestFunctionThrowsExceptionsAndReturns(t *testing.T) {
	reports := linttest.GetFileReports(t, `<?php
	class Exception {}

	function handle($b) {
		if ($b === 1) {
			return $b;
		}

		switch ($b) {
			case "a":
				throw new \Exception("a");

			default:
				throw new \Exception("default");
		}
	}

	function doSomething() {
		handle(1);
		echo "This code is reachable\n";
	}`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	fi, ok := meta.Info.GetFunction(`\handle`)

	if ok {
		log.Printf("handle exitFlags: %d (%s)", fi.ExitFlags, linter.FlagsToString(fi.ExitFlags))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestRedundantCast(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function bad($a) {
		$int = 1;
		$double = 1.0;
		$string = '1';
		$bool = ($a == 0);
		$array = [1, 'a', 3.0]; // Mixed elems on purpose
		$a = (int)$int;
		$a = (double)$double;
		$a = (string)$string;
		$a = (bool)$bool;
		$a = (array)$array;
		$_ = $a;
	}

	function good($a) {
		$int = 1;
		$double = 1.0;
		$string = '1';
		$bool = ($a == 0);
		$array = [1, 'a', 3.0];
		$a = (int)$double;
		$a = (double)$array;
		$a = (string)$bool;
		$a = (bool)$string;
		$a = (array)$int;
		$_ = $a;
	}`)
	test.Expect = []string{
		`expression already has array type`,
		`expression already has float type`,
		`expression already has int type`,
		`expression already has string type`,
		`expression already has bool type`,
	}
	test.RunAndMatch()
}

func TestSwitchBreak(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function bad($a) {
		switch ($a) {
		case 1:
			echo "One\n"; // Bad, no break.
		default:
			echo "Other\n";
		}
	}

	function good($a) {
		switch ($a) {
		case 1:
			echo "One\n";
			break;
		case 2:
			echo "Two";
			// No break, but still good, since it's the last case clause.
		}

		echo "Three";
	}`)
	test.Expect = []string{`Add break or '// fallthrough' to the end of the case`}
	test.RunAndMatch()
}

func TestNameCase(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class FooBar {
  public function method_a() {}
}

class Baz extends foobar {}

$foo = new Foobar();
$foo->Method_a();

function func_a() {}

func_A();
`)
	test.Expect = []string{
		`\Foobar should be spelled \FooBar`,
		`\foobar should be spelled \FooBar`,
		`Method_a should be spelled method_a`,
		`\func_A should be spelled \func_a`,
	}
	runFilterMatch(test, `nameCase`)
}

func TestClassNotFound(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$_ = new Foo();

class Derived extends Base {}

class Impl implements Iface1, Iface2 {}

interface Iface extends IfaceBase {}
`)
	test.Expect = []string{
		`Type \Base not found`,
		`Type \Iface1 not found`,
		`Type \Iface2 not found`,
		`Type \Foo not found`,
	}
	test.RunAndMatch()
}

func TestCorrectArrayTypes(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function test() {
		$a = [ 'a' => 123, 'b' => 3456 ];
		return $a['a'];
	}
	`)
	test.RunLinter()

	fn, ok := meta.Info.GetFunction(`\test`)
	if !ok {
		t.Errorf("Could not find function test")
		t.Fail()
	}

	if l := fn.Typ.Len(); l != 1 {
		t.Errorf("Unexpected number of types: %d, excepted 1", l)
	}

	if !fn.Typ.Is("int") {
		t.Errorf("Wrong type: %s, expected int", fn.Typ)
	}
}

func TestArrayUnion(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function testInt() {
		return 1 + 1;
	}
	function testIntArr() {
		return [1] + [2];
	}
	function testMixedArr() {
		return [1] + ['foo'];
	}
	`)
	test.RunLinter()

	fnInt, ok := meta.Info.GetFunction(`\testInt`)
	if !ok {
		t.Errorf("Could not find function testInt")
		t.Fail()
	}

	if l := fnInt.Typ.Len(); l != 1 {
		t.Errorf("Unexpected number of types: %d, excepted 1", l)
	}

	if !fnInt.Typ.Is("int") {
		t.Errorf("Wrong type: %s, expected int", fnInt.Typ)
	}

	fnIntArr, ok := meta.Info.GetFunction(`\testIntArr`)
	if !ok {
		t.Errorf("Could not find function testIntArr")
		t.Fail()
	}

	if l := fnIntArr.Typ.Len(); l != 1 {
		t.Errorf("Unexpected number of types: %d, excepted 1", l)
	}

	if !fnIntArr.Typ.IsArrayOf("int") {
		t.Errorf("Wrong type: %s, expected int[]", fnIntArr.Typ)
	}

	fnMixedArr, ok := meta.Info.GetFunction(`\testMixedArr`)
	if !ok {
		t.Errorf("Could not find function testMixedArr")
		t.Fail()
	}

	if l := fnMixedArr.Typ.Len(); l != 2 {
		t.Errorf("Unexpected number of types: %d, excepted 2", l)
	}

	if !fnMixedArr.Typ.Equals(meta.NewTypesMap("int[]|string[]")) {
		// NOTE: this is how code works right now. It currently treat a[]|b[] as (a|b)[]
		t.Errorf("Wrong type: %s, expected int[]|string[]", fnMixedArr.Typ)
	}
}

func TestCompactImpliesUsage(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function define($_, $_) {}
define('null', 0);

// Declaration from phpstorm-stubs
function compact ($varname, $_ = null) {}

function f() {
	$x = 1; $y = 2;
	// Equivalent to ['x' => $x, 'y' => $y]
	return compact('x', 'y');
}

function g() {
	$x = 1; $y = 2;
	// Also equivalent to ['x' => $x, 'y' => $y]
	return compact([[['x'], 'y']]);
}
	`)
}

func TestCompactWithUndefined(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function define($_, $_) {}
define('null', 0);

// Declaration from phpstorm-stubs
function compact ($varname, $_ = null) {}

function f() {
	return compact('x', 'y');
}
	`)

	test.Expect = []string{
		"Undefined variable: x",
		"Undefined variable: y",
	}
	runFilterMatch(test, "undefined")
}

func TestAssignByRef(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function a() {
	  $a = 1;
	  $b = &$a;
	  $b = 2;
	  return $a;
	}

	echo a();`)
}

func TestUndefinedConst(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
echo UNDEFINED_CONST;
`)
	test.Expect = []string{`Undefined constant UNDEFINED_CONST`}
	test.RunAndMatch()
}

func addNamedFile(test *linttest.Suite, name, code string) {
	test.Files = append(test.Files, linttest.TestFile{
		Name: name,
		Data: []byte(code),
	})
}

func runFilterMatch(test *linttest.Suite, names ...string) {
	test.Match(filterReports(names, test.RunLinter()))
}

func filterReports(names []string, reports []*linter.Report) []*linter.Report {
	set := make(map[string]struct{})
	for _, name := range names {
		set[name] = struct{}{}
	}

	var out []*linter.Report
	for _, r := range reports {
		if _, ok := set[r.CheckName()]; ok {
			out = append(out, r)
		}
	}
	return out
}
