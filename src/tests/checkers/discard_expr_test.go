package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestDiscardExprInternalClassCall(t *testing.T) {
	test := linttest.NewSuite(t)
	test.LoadStubs = []string{
		`stubs/phpstorm-stubs/dom/dom_c.php`,
	}
	test.AddFile(`<?php
class MyDoc extends DOMDocument {
  /**/
  public function loadContent($content) {
    $this->loadHTML('<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">' . $content);
  }
}

$content = '';
$dom = new DOMDocument();
$dom->loadHTML('<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">' . $content);
`)
	linttest.RunFilterMatch(test, "discardExpr")
}

func TestDiscardExprVariableCall(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function local1($xs) {
  // It does discard the result, but we're not tracking
  // variable function calls yet.
  $process = function($x) { return $x; };
  foreach ($xs as $x) {
    $process($x);
  }
}

function local2($xs) {
  $process = function($x) { exit(0); };
  foreach ($xs as $x) {
    echo 123;
    $process($x);
  }
}

function local3($xs) {
  $process = function($x) { echo $x; };
  $process($xs);
}

function local4($xs) {
  $process = function($x) { return $x; };
  $process($xs);
}
`)
}

func TestDiscardExprAbstractCall(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
abstract class TheAbstractClass {
  /** @return bool */
  abstract public function doIt();
}

class TheClass extends TheAbstractClass {
  /** @return bool */
  public function doIt() { return (bool)1; }
}

function useAbstract(TheAbstractClass $ac) {
  $ac->doIt();
}

$c = new TheClass();

useAbstract($c);
`)
}

func TestDiscardExprInterfaceCall(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
interface TheInterface {
  /** @return bool */
  public function load($x);
}

class TheClass implements TheInterface {
  /** @return bool */
  public function load($x) { return (bool)0; }
}

function useInterface(TheInterface $i) {
  $i->load(10);
}

$c = new TheClass();

useInterface($c);
`)
}

func TestDiscardExprCall(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
define('null', 0);

function count($xs) { return 0; }

class C {
  public function pure1() {
    return 10;
  }

  public function pure2($xs) {
    return count($xs) + 1;
  }

  public static function create() {
    return new static();
  }

  public function notPure1() {
    throw new Exception("123");
  }
  public function notPure2() {
    exit(0);
  }
  public function notPure3() {
    $this->notPure2();
  }
}

function pure_fn1($x, $y) {
  if ($x && $y) {
    return null;
  }
  return [$x, $y];
}

$fn = function() { return 10; };
$fn();

$c = new C();
$c->notPure1();
$c->notPure2();
$c->notPure3();
$c->pure1(); // warn 1
$c->pure2(); // warn 2
C::create(); // warn 3
pure_fn1(1, 2); // warn 4
`)
	test.Expect = []string{
		`expression evaluated but not used`,
		`expression evaluated but not used`,
		`expression evaluated but not used`,
		`expression evaluated but not used`,
	}
	linttest.RunFilterMatch(test, "discardExpr")
}

func TestDiscardExpr(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class C {}

function count($xs) { return 0; }

$xs = [];

count($xs); // Unused count()

function f() {
  1; // Unused literal
}

[1, 2]; // Unused array literal
new C(); // Unused new expression

1 + 4; // Unused binary expr
count($xs) * 2; // Unused binary expr

$xs /*::array<int>*/;

class Foo {
  private static $x;

  private function f() {
    self::$x /*::int*/;
    return self::$x;
  }
}

$a = 10;
$xs ??= $a; // Ok
`)
	test.Expect = []string{
		`expression evaluated but not used`,
		`expression evaluated but not used`,
		`expression evaluated but not used`,
		`expression evaluated but not used`,
		`expression evaluated but not used`,
		`expression evaluated but not used`,
		`expression evaluated but not used`,
		`expression evaluated but not used`,
	}
	test.RunAndMatch()
}
