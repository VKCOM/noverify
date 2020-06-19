package linttest_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue327(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function sink(...$args) {}

function f() {
  return <<<'SQL'
    SELECT login, password FROM user WHERE login LIKE '%admin%'
  SQL;
}

function f2($x) {
  sink(<<<STR
    $x $x
    STR, 1, 2);

  sink(<<<STR
    abc
  STR);

  sink(<<<"STR"
  STRNOTEND
    abc
  STR);
}
`)
}

func TestIssue289(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo { public $value = 11; }

$xs = [0, new Foo()];

/* @var Foo $foo */
$foo = $xs[1];
$_ = $foo->value;
`)
}

func TestIssue1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	interface TestInterface
	{
		const TEST = '1';
	}

	class TestClass implements TestInterface
	{
		/** get returns interface constant */
		public function get()
		{
			return self::TEST;
		}
	}`)
}
func TestIssue2(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	function rand() { return 4; }

	interface DateTimeInterface {
		public function format($fmt);
	}

	interface TestClassInterface
	{
		public function getCreatedAt(): \DateTimeInterface;
	}

	function test(): \DateTimeInterface {
		return 0; // this should return error as well :)
	}

	function a(TestClassInterface $testClass): string
	{
		if (rand()) {
			return $testClass->getCreatedAt()->format('U');
		} else {
			return test()->format('U');
		}
	}`)
}

func TestIssue3(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class TestClass
	{
		/** get always returns "." */
		public function get(): string
		{
			return '.';
		}
	}

	function a(TestClass ...$testClasses): string
	{
		$result = '';
		foreach ($testClasses as $testClass) {
			$result .= $testClass->get();
		}

		return $result;
	}

	echo a(new TestClass()), "\n";
	echo a(); // OK to call with 0 arguments.
	`)
}

func TestIssue6(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	declare(strict_types=1);

	trait Example
	{
		private static $property = 'some';

		protected function some(): string
		{
			return self::$property;
		}
	}`)
}

func TestIssue8(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class Magic
	{
		public function __get();
		public function __set();
		public function __call();
	}

	class MagicStatic {
		public static function __callStatic();
	}

	function test() {
		$m = new Magic;
		echo $m->some_property;
		$m->another_property = 3;
		$m->call_something();
		MagicStatic::callSomethingStatic();
	}`)
}

func TestIssue11(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class Generator {
		/** send sends a message */
		public function send();
	}

	function a($a): \Generator
	{
		yield $a;
	}

	a(42)->send(42);
	`)
}
func TestIssue16(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	interface DateTimeInterface {
		public function format($fmt);
	}

	interface OtherInterface {
		public function useless();
	}

	interface TestInterface
	{
		const TEST = 1;

		public function getCreatedAt(): \DateTimeInterface;
	}

	interface TestExInterface extends OtherInterface, TestInterface
	{
	}

	function a(TestExInterface $testInterface): string
	{
		echo TestExInterface::TEST;
		return $testInterface->getCreatedAt()->format('U');
	}

	function b(TestExInterface $testInterface) {
		echo TestExInterface::TEST2;
		return $testInterface->nonexistent()->format('U');
	}`)
	test.Expect = []string{
		`Call to undefined method {\TestExInterface}->nonexistent()`,
		"Call to undefined method {mixed}->format()",
		"Class constant \\TestExInterface::TEST2 does not exist",
	}
	runFilterMatch(test, "undefined")
}

func TestIssue26_1(t *testing.T) {
	// Test that defined variable variable don't cause "undefined" warnings.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		$x = 'key';
		if (isset($$x)) {
			$_ = $x + 1;  // If $$x is isset, then $x is set as well
			$_ = $$x + 1;
			$_ = $y;      // Undefined
		}
		// After the block all vars are undefined again.
		$_ = $x;
	}`)
	test.Expect = []string{"Undefined variable: y"}
	test.RunAndMatch()
}

func TestIssue26_2(t *testing.T) {
	// Test that if $x is defined, it doesn't make $$x defined.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($x)) {
			$_ = $x;  // $x is defined
			$_ = $$x; // But $$x is not
		}
	}`)
	test.Expect = []string{"Unknown variable variable $$x used"}
	test.RunAndMatch()
}

func TestIssue26_3(t *testing.T) {
	// Test that irrelevant isset of variable-variable doesn't affect
	// other variables. Also warn for undefined variable in $$x.
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($$x)) {
			$_ = $$y;
		}
	}`)
	test.Expect = []string{
		"Undefined variable: x",
		"Unknown variable variable $$y used",
	}
	test.RunAndMatch()
}

func TestIssue26_4(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	function issetVarVar() {
		if (isset($$$$x)) {
			$_ = $$$$x; // Can't track this level of indirection
		}
	}`)
	test.Expect = []string{
		"Unknown variable variable $$$x used",
		"Unknown variable variable $$$$x used",
	}
	test.RunAndMatch()
}

func TestIssue37(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class Foo {
		public $a;
		public $b;
	}

	/**
	 * @param Foo[] $arr
	 */
	function f($arr) {
		$ads_ids = array_keys($arr);
		foreach ($ads_ids as $num => $ad_id) {
			if ($num + 1 < count($ads_ids)) {
				$second_ad_id = $ads_ids[$num + 1];
				$arr[$ad_id]->a = $arr[$second_ad_id]->b;
			}
		}
	}`)
	runFilterMatch(test, "unused")
}

func TestIssue78_1(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
global $cond;
$xs = [1, 2];
switch ($cond) {
case 0:
  trailing_exit_if($xs);
  echo "unreachable";
  break;
case 1:
  trailing_exit_foreach($xs);
  echo "unreachable";
  break;
case 2:
  trailing_exit_foreach2($xs);
  echo "unreachable";
  break;
case 3:
  trailing_throw_if($xs);
  echo "unreachable";
  break;
case 4:
  trailing_throw_foreach($xs);
  echo "unreachable";
  break;
case 5:
  trailing_throw_foreach2($xs);
  echo "unreachable";
  break;
case 6:
  trailing_exit_for($xs);
  echo "unreachable";
  break;
case 7:
  trailing_exit_while($xs);
  echo "unreachable";
  break;
case 8:
  trailing_exit_try($xs);
  echo "unreachable";
  break;
case 9:
  trailing_exit_try2($xs);
  echo "unreachable";
  break;
case 10:
  trailing_exit_catch($xs);
  echo "unreachable";
  break;
case 11:
  trailing_exit_switch($xs);
  echo "unreachable";
  break;
}

class Exception {}

function trailing_exit_switch($xs) {
  switch($xs[0]) {
  case 1:
    die("ok");
  }
  exit;
}

function trailing_exit_try($xs) {
  try {
    if ($xs) {
      die("ok");
    }
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_try2($xs) {
  try {
    try {
      if ($xs) {
        if ($xs[0] < 1000) {
          die("ok");
        }
      }
    } catch (Exception $_) {}
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_catch($xs) {
  try {
  } catch (Exception $_) {
    die("ok");
  }
  exit;
}

function trailing_exit_if($xs) {
  if ($xs) {
    die("ok");
  }
  exit;
}

function trailing_exit_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
      die("ok");
    }
  }
  exit;
}

function trailing_exit_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
        die("ok");
      }
    }
  }
  exit;
}

function trailing_throw_if($xs) {
  if ($xs) {
    die("ok");
  }
  throw new Exception("oops");
}

function trailing_throw_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
      die("ok");
    }
  }
  throw new Exception("oops");
}

function trailing_throw_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
        die("ok");
      }
    }
  }
  throw new Exception("oops");
}

function trailing_exit_for($xs) {
  for ($i = 0; $i < 10; $i++) {
    if ($i == $xs[0]) {
      die("ok");
    }
  }
  exit;
}

function trailing_exit_while($xs) {
  while (1) {
    if ($xs[0] < 1000) {
      die("ok");
    }
    break;
  }
  exit;
}`)

	test.Expect = []string{
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
	}

	test.RunAndMatch()
}

func TestIssue78_2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
global $cond;
$xs = [1, 2];
switch ($cond) {
case 0:
  trailing_exit_if($xs);
  echo "unreachable";
  break;
case 1:
  trailing_exit_foreach($xs);
  echo "unreachable";
  break;
case 2:
  trailing_exit_foreach2($xs);
  echo "unreachable";
  break;
case 3:
  trailing_throw_if($xs);
  echo "unreachable";
  break;
case 4:
  trailing_throw_foreach($xs);
  echo "unreachable";
  break;
case 5:
  trailing_throw_foreach2($xs);
  echo "unreachable";
  break;
case 6:
  trailing_exit_for($xs);
  echo "unreachable";
  break;
case 7:
  trailing_exit_while($xs);
  echo "unreachable";
  break;
case 8:
  trailing_exit_try($xs);
  echo "unreachable";
  break;
case 9:
  trailing_exit_try2($xs);
  echo "unreachable";
  break;
case 10:
  trailing_exit_catch($xs);
  echo "unreachable";
  break;
case 11:
  trailing_exit_switch($xs);
  echo "unreachable";
  break;
}

class Exception {}

function trailing_exit_switch($xs) {
  switch($xs[0]) {
  case 1:
    $_ = $xs[0];
  }
  exit;
}

function trailing_exit_try($xs) {
  try {
    if ($xs) {
    }
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_try2($xs) {
  try {
    try {
      if ($xs) {
        if ($xs[0] < 1000) {
        }
      }
    } catch (Exception $_) {}
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_catch($xs) {
  try {
  } catch (Exception $_) {
  }
  exit;
}

function trailing_exit_if($xs) {
  if ($xs) {
  }
  exit;
}

function trailing_exit_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
    }
  }
  exit;
}

function trailing_exit_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
      }
    }
  }
  exit;
}

function trailing_throw_if($xs) {
  if ($xs) {
  }
  throw new Exception("oops");
}

function trailing_throw_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
    }
  }
  throw new Exception("oops");
}

function trailing_throw_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
      }
    }
  }
  throw new Exception("oops");
}

function trailing_exit_for($xs) {
  for ($i = 0; $i < 10; $i++) {
    if ($i == $xs[0]) {
    }
  }
  exit;
}

function trailing_exit_while($xs) {
  while (1) {
    if ($xs[0] < 1000) {
    }
  }
  exit;
}`)

	test.Expect = []string{
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
		"Unreachable code",
	}

	test.RunAndMatch()
}

func TestIssue78_3(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
$xs = [1, 2];
trailing_exit_if($xs);
trailing_exit_foreach($xs);
trailing_exit_foreach2($xs);
trailing_throw_if($xs);
trailing_throw_foreach($xs);
trailing_throw_foreach2($xs);
trailing_exit_for($xs);
trailing_exit_while($xs);
trailing_exit_try($xs);
trailing_exit_try2($xs);
trailing_exit_catch($xs);
trailing_exit_switch($xs);
echo "not a dead code";

class Exception {}

function trailing_exit_switch($xs) {
  switch($xs[0]) {
  case 1:
    return "ok";
  }
  exit;
}

function trailing_exit_try($xs) {
  try {
    if ($xs) {
      return "ok";
    }
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_try2($xs) {
  try {
    try {
      if ($xs) {
        if ($xs[0] < 1000) {
          return "ok";
        }
      }
    } catch (Exception $_) {}
  } catch (Exception $_) {}
  exit;
}

function trailing_exit_catch($xs) {
  try {
  } catch (Exception $_) {
    return "ok";
  }
  exit;
}

function trailing_exit_if($xs) {
  if ($xs) {
    return "ok";
  }
  exit;
}

function trailing_exit_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
      return "ok";
    }
  }
  exit;
}

function trailing_exit_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
        return "ok";
      }
    }
  }
  exit;
}

function trailing_throw_if($xs) {
  if ($xs) {
    return "ok";
  }
  throw new Exception("oops");
}

function trailing_throw_foreach($xs) {
  foreach ($xs as $x) {
    if ($x < 10) {
      return "ok";
    }
  }
  throw new Exception("oops");
}

function trailing_throw_foreach2($xs) {
  foreach ([$xs] as $ys) {
    foreach ($ys as $y) {
      if ($y < 10) {
        return "ok";
      }
    }
  }
  throw new Exception("oops");
}

function trailing_exit_for($xs) {
  for ($i = 0; $i < 10; $i++) {
    if ($i == $xs[0]) {
      return "ok";
    }
  }
  exit;
}

function trailing_exit_while($xs) {
  while (1) {
    if ($xs[0] < 1000) {
      return "ok";
    }
    break;
  }
  exit;
}
`)
}

func TestIssue128(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Value {
  public $x;
}

function count($arr) { return 0; }

function good($v) {
  if (isset($good) && count($good) == 1) {}

  if ($v instanceof Value && $v->x) {}
  if (isset($y) && $y instanceof Value && $y->x) {}
}

function bad1($v) {
  if (isset($bad0) && $bad0) {}
  $_ = $bad0; // Used outside of if body

  if (count($bad1) == 1 && isset($bad1)) {}
  if (isset($good) && count($bad2) == 1 && isset($bad2)) {}
  if (isset($bad3) || count($bad3) == 1) {}

  if ($v->x && $v instanceof Value) {}

  if ($y1 instanceof Value && isset($y1) && $y1->x) {}
}

$_ = $bad1;
`)
	test.Expect = []string{
		`Undefined variable: bad0`,
		`Undefined variable: bad1`, // At local scope
		`Undefined variable: bad1`, // At global scope
		`Undefined variable: bad2`,
		`Undefined variable: bad3`,
		`Property {mixed}->x does not exist`,
		`Variable might have not been defined: y1`,
	}
	test.RunAndMatch()
}

func TestIssue170(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php

global $v;

switch ($v) {
case 1:
  error(); // no break here
case 2:
  $_ = $v;
  break;
}

function error() {}
`)
}

func TestIssue182(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
define('null', 0);

function define($name, $v) {}

trait SingletonSelf {
    /** @var self */
    private static $instance = null;

    /** @return self */
    public static function instance() {
        if (self::$instance === null) {
            self::$instance = new self();
        }

        return self::$instance;
    }
}

trait SingletonStatic {
    /** @var static */
    private static $instance = null;

    /** @return static */
    public static function instance() {
        if (static::$instance === null) {
            static::$instance = new static();
        }

        return static::$instance;
    }
}
`)
}

func TestIssue183(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
    trait Mixin {
        public $x = 10;
    }

    class MyClass {
        use Mixin;

        /** @return int */
        public function useX() { return $this->x; }
        /** @return int */
        public function useY() { return $this->y; }
    }
`)

	test.Expect = []string{
		`Property {\MyClass}->y does not exist`,
	}

	test.RunAndMatch()
}

func TestIssue362_1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function method_exists($object, $method_name) { return 1 != 0; }

class Foo {
  public $value;
}

$x = new Foo();
if (method_exists($x, 'm1')) {
  $x->m1();
  $x->m1(1, 2);
}

if (method_exists(new Foo(), 'm2')) {
  if (method_exists(new Foo(), 'm3')) {
    (new Foo())->m2((new Foo())->m3());
  }
  (new Foo())->m2();
}

$y = new Foo();
if (method_exists($x, 'x1')) {
  $x->x1();
} elseif (method_exists($y, 'y1')) {
  $foo = $y->y1();
  if ($foo instanceof Foo) {
    echo $foo->value;
  }
}
`)
}

func TestIssue362_2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function method_exists($object, $method_name) { return 1 != 0; }

class Foo {}

$x = new Foo();
if (method_exists($x, 'm1')) {
}
$x->m1(); // Bad: called outside of if

if (method_exists(new Foo(), 'm2')) {
  if (method_exists(new Foo(), 'm3')) {
    $x->m2(); // Bad: called a method on a different object expression
  }
}
(new Foo())->m3(); // Bad: called outside of if

$y = new Foo();
if (method_exists($x, 'x1')) {
  $x->y1();
} elseif (method_exists($y, 'y1')) {
  $v = $y->x1();
  $v->foo();
}
`)
	test.Expect = []string{
		`Call to undefined method {\Foo}->m1()`,
		`Call to undefined method {\Foo}->m2()`,
		`Call to undefined method {\Foo}->m3()`,
		`Call to undefined method {\Foo}->y1()`,
		`Call to undefined method {\Foo}->x1()`,
		`Call to undefined method {mixed}->foo()`,
	}
	test.RunAndMatch()
}

func TestIssue252(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {
  public $foo = 10;
}
class Bar {
  public $bar = 20;
}
function alt_foreach($arr) {
  foreach ($arr AS $key => $value):
    $_ = [$key, $value];
  endforeach;
}
function alt_if($v) {
  if ($v instanceof Foo):
    $_ = $v->foo;
  elseif ($v instanceof Bar):
    $_ = $v->bar;
  endif;
}
`)

	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function alt_for() {
  for ($i = 0; $i < 10; $i++):
    $x1 = 10;
  endfor;
  $_ = $x1;
}
function alt_switch($v) {
  switch ($v):
  case 1:
    $v = 3;
  case 2:
    return $v;
  endswitch;
}`)
	test.Expect = []string{
		`Variable might have not been defined: x1`,
		`Add break or '// fallthrough' to the end of the case`,
	}
	test.RunAndMatch()
}

func TestIssue387(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function f1(&$a) {
    $a1[0] = 1;
}

// TODO: report that $result is unchanged even if we tried to
// modify its arguments through references? See #388
function f2($result) {
    foreach ($result as &$file) {
        $file['filesystem'] = 'ntfs';
    }

    // Should be reported by #388.
    $arr = [];
    $arr['x'] = 10;
}
`)
}

func TestIssue390(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$cond = 1;
if ($cond && isset($a1[0])) {
  $_ = $a1;
}
if ($cond && isset($a2[0][1])) {
  $_ = $a2;
}
if (isset($a3[0]) && $cond) {
  $_ = $a3;
}
if (isset($a4[0][1]) && $cond) {
  $_ = $a4;
}

if (isset($a5[0]->x)) {
  $_ = $a5;
}
`)
	test.Expect = []string{
		`Property {mixed}->x does not exist`,
	}
	test.RunAndMatch()
}

func TestIssue288(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Box {
  public $item1;
  public $item2;
}

$_ = isset($a) ? $a[0] : 0;
$_ = isset($b) && isset($a) ? $a[0] + $b : 0;
$_ = isset($a[0]) ? $a[0] : 0;
$_ = isset($a[0]) ? $a : [0];

$_ = isset($b1) ? 0 : $b1;
$_ = isset($b2[0]) ? 0 : $b2;
$_ = isset($b3[0]) ? 0 : $b3;


function f($x, $y) {
  $_ = $x instanceof Box ? $x->item1 : 0;
  $_ = $y instanceof Box ? 0 : $y->item2;
}

$x = new Box();
$_ = $badvar ? 0 : 1;
$_ = isset($x) && isset($y) ? $x : 0;
$_ = $x instanceof Box ? 0 : 1;
`)
	test.Expect = []string{
		`Undefined variable: badvar`,
		`Undefined variable: b1`,
		`Undefined variable: b2`,
		`Undefined variable: b3`,
		`undefined: Property {mixed}->item2 does not exist`,
	}
	test.RunAndMatch()
}

func TestIssue283(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait YummyProps {
  public $price = 'fair';
  protected $taste = 'good';
  private $secret = 'sauce';
}

class Borsch {
  use YummyProps;

  /** @return string */
  public function getTaste() {
    return $this->taste; // OK: using protected trait prop from embedding class
  }

  /** @return string */
  public function getSecret() {
    return $this->secret; // OK: using private trait prop from embedding class
  }
}

class SaltyBorsch extends Borsch {
  protected function getTaste2() {
    return $this->taste; // OK: using protected trait prop from inherited class
  }
  protected function getSecret2() {
    return $this->secret; // Bad: used private trait prop from inherited class
  }
}

$borsch = new Borsch();
$_ = $borsch->taste; // Bad: can't access protected trait prop from the outside
$_ = $borsch->price; // OK: can use public trait prop here

$borsch2 = new SaltyBorsch();
$_ = $borsch2->price; // OK: can also access public trait prop from derived class
`)
	test.Expect = []string{
		`Cannot access private property \Borsch->secret`,
		`Cannot access protected property \Borsch->taste`,
	}
	test.RunAndMatch()
}

func TestIssue209_1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
trait A {
  private function priv() { return 1; }
  protected function prot() { return 2; }
  /** @return int */
  public function pub() { return 3; }
}

class B {
  use A;
  /** @return int */
  public function sum() {
    return $this->priv() + $this->prot() + $this->pub();
  }
}

echo (new B)->sum(); // actual PHP prints 6
`)
}

func TestIssue209_2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait Methods {
  /***/
  public function pubMethod() {}
  protected function protMethod() {}
  private function privMethod() {}

  /***/
  public static function staticPubMethod() {}
  protected static function staticProtMethod() {}
  private static function staticPrivMethod() {}
}

class Base {
  use Methods;

  private function f() {
    $this->pubMethod();
    $this->protMethod();
    $this->privMethod();
  }
}

class Derived extends Base {
  private function f() {
    $this->pubMethod();
    $this->protMethod();
    $this->privMethod(); // Bad: can't call private
    Base::staticPubMethod();
    Base::staticProtMethod();
    parent::staticProtMethod();
    parent::staticPrivMethod(); // Bad: can't call private
  }
}

$b = new Base();
$b->pubMethod();
$b->protMethod(); // Bad: can't call from the outside
$b->privMethod(); // Bad: can't call from the outside

Base::staticPubMethod();
Derived::staticPubMethod();
Base::staticProtMethod();
`)
	test.Expect = []string{
		`Cannot access private method \Base->privMethod()`,
		`Cannot access protected method \Base->protMethod()`,
		`Cannot access private method \Base->privMethod()`,
		`Cannot access private method \Base::staticPrivMethod()`,
		`Cannot access protected method \Base::staticProtMethod()`,
	}
	test.RunAndMatch()
}

func TestIssue497(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
/**
 * @param shape(a:int) $x
 * @return T<int>
 */
function f($x) {
  $v = $x['a'];
  return [$v];
}
`)
}

func TestDupArrayKeys_ToSkip(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`
  <?php
  class T {}

  $skips_1 = [
    new T() => 1,
    new T() => 2,
  ];

  $skips_2 = [
    [1, 2] => 1,
    [1, 2] => 2,
    [1, 2,] => 3,
    [1, 2,] => 4,
  ];

  $skips_3 = [
    function() {} => 1,
    function() {} => 2,
  ];
  ?>
  `)
}

func TestDupArrayKeys_Consts(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`
  <?php
  $C1 = 1;
  $C2 = 2;
  const c1 = 1;
  const c2 = 2;

  class T {
    const C1 = 1;
    const C2 = 2;
  }

  $constants_1 = [
    $C1 => 1, 
    $C2 => 2, 
    $C1 => 3, // Duplicate key C1
  ];

  $constants_2 = [
    T::C1 => 1, 
    T::C2 => 2, 
    T::C1 => 3, // Duplicate key T1::C1
  ];

  $constants_3 = [
    c1 => 1,
    c2 => 2,
    c1 => 3, // Duplicate key c1
  ];
  ?>
  `)

	test.Expect = []string{
		`Duplicate array key '$C1'`,
		`Duplicate array key 'T::C1'`,
		`Duplicate array key 'c1'`,
	}

	test.RunAndMatch()
}

func TestDupArrayKeys_Nums(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`
  <?php
  $nums = [
    73 => 1, 
    2 => 2, 
    73 => 3, // Duplicate key 73
    73.0 => 4, // Duplicate key 73.0
    0b1001001 => 5, // Duplicate key 0b1001001
    0x49 => 6, // Duplicate key 0x49
    7_3 => 7, // Duplicate key 7_3
    0111 => 8, // Duplicate key 0111
    "73" => 9, // Duplicate key 73
    0.73e2 => 10, // Duplicate key 0.73e2
  ];
  ?>
  `)

	test.Expect = []string{
		`Duplicate array key '73'`,
		`Duplicate array key '73.0'`,
		`Duplicate array key '0b1001001'`,
		`Duplicate array key '0x49'`,
		`Duplicate array key '7_3'`,
		`Duplicate array key '0111'`,
		`Duplicate array key '"73"'`,
		`Duplicate array key '0.73e2'`,
	}

	test.RunAndMatch()
}

func TestDupArrayKeys_Strings(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`
  <?php
  $first = "a";
  const first = "b";

  $strings_1 = [
    "first" => 1,
    "second" => 2,
    "first" => 3, // Duplicate key first
    $first => 4,
    first => 5,
  ];

  // Heredocs
  $strings_2 = [
  <<<EOT
1
EOT => 1,
  <<<EOT
2
EOT => 2,
  <<<EOT
1
EOT => 3,
]; // Duplicate key 
  ?>
  `)

	test.Expect = []string{
		`Duplicate array key '"first"'`,
		`Duplicate array key '<<<EOT
1
EOT'`,
	}

	test.RunAndMatch()
}

func TestDupArrayKeys_Expressions(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`
  <?php
  $s = '123';

  function getX($x) {
    return $x;
  }

  $someVar = 10;
  function getFluidVar() {
    global $someVar;
    $someVar += 10;
    return $someVar;
  }

  $k = [11, 22];

  class T {}

  $expressions_1 = [
    $k[0] => 1,
    $k[1] => 2,
    $k[0] => 3, // Duplicate key $k[0]
  ];

  // Const pure function
  $expressions_2 = [
    getX(1) => 1,
    getX(2) => 2,
    getX(1) => 3, // Duplicate key getX(1)
  ];

  // It's ok
  $expressions_3 = [
    getFluidVar() => 1,
    getFluidVar() => 2,
    getFluidVar() => 3,
  ];

  // Built-in functions
  $expressions_4 = [
    isset($k) => 1,
    isset($k) => 2, // Duplicate key isset($k)
  ];

  // Built-in operators
  $expressions_5 = [
    $k instanceof T => 1,
    $k instanceof T => 2, // Duplicate key $k instanceof T
  ];
  
  // Concat
  $expressions_6 = [
    'a' . $s => 1,
    'b' . $s => 2,
    'a' . $s => 3, // Duplicate key 'a'.$s
  ];

  $a = 1;
  $b = 2;
  $expressions_7 = [
    $a => 1,
    $a + 1 => 2,
    $a + 1 => 3, // Duplicate key $a + 1
    $a + $b => 4,
    $a + $b => 5, // Duplicate key $a + $b
  ];
  ?>
  `)

	test.Expect = []string{
		`Duplicate array key '$k[0]'`,
		`Duplicate array key 'getX(1)'`,
		`Duplicate array key 'isset($k)'`,
		`Duplicate array key '$k instanceof T'`,
		`Duplicate array key ''a' . $s'`,
		`Duplicate array key '$a + 1'`,
		`Duplicate array key '$a + $b'`,
	}

	test.RunAndMatch()
}
