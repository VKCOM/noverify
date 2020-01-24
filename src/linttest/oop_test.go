package linttest_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestOldStyleConstructor(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class T1 {
  /** simple constructor */
  public function T1() {}
}

class T2 {
  /** constructor name is in lower case */
  public function t2() {}
}

class t3 {
  /** inverse of the T2 test case */
  public function T3() {}
}
`)
	test.Expect = []string{
		`Old-style constructor usage, use __construct instead`,
		`Old-style constructor usage, use __construct instead`,
		`Old-style constructor usage, use __construct instead`,
	}
	test.RunAndMatch()
}

func TestConstructorArgCount(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
use A\B;

class OneArg {
  public function __construct($_) {}
}

class OneArgDerived extends OneArg {}

function f() {
  $_ = new OneArg();
  $_ = new \A\B\TwoArgs();
  $_ = new B\TwoArgs;
  $_ = new OneArgDerived();
}
`)
	test.AddFile(`<?php
namespace A\B;

class TwoArgs {
  public function __construct($_, $_) {}
}
`)
	test.Expect = []string{
		`Too few arguments for \OneArg constructor`,
		`Too few arguments for \A\B\TwoArgs constructor`,
		`Too few arguments for \A\B\TwoArgs constructor`,
		`Too few arguments for \OneArgDerived constructor`,
	}
	test.RunAndMatch()
}

func TestPhpdocProperty(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * @property int $int
 * @property string $string - optional description.
 */
class WithProps {
  /***/
  public function getInt() {
    return $this->int;
  }
  /***/
  public function getString() {
    return $this->string;
  }
}

$_ = (new WithProps())->int;
$_ = (new WithProps())->string;

function f(WithProps $x) {
  return $x->int + $x->int;
}

// Can't access them as static props/constants.
$_ = WithProps::int;
$_ = WithProps::$int;
`)

	test.Expect = []string{
		`Class constant \WithProps::int does not exist`,
		`Property \WithProps::$int does not exist`,
	}

	test.RunAndMatch()
}

func TestInheritDoc(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
interface TestInterface {
  /**
   * @return self
   */
  public function getThis();

  /**
   * @param \TestInterface $x
   */
  public function acceptThis($x);

  /**
   * @param mixed $p1
   * @param mixed $p2
   * @param mixed $p3
   * @param mixed $p4
   * @param \TestInterface $x
   */
  public function acceptThis5($p1, $p2, $p3, $p4, $x);
}
`)
	test.AddFile(`<?php
class Foo implements TestInterface {
  /** @inheritdoc */
  public function getThis() { return $this; }

  /** @inheritdoc */
  public function acceptThis($x) { return $x->getThis(); }

  /** Should work regardless of "inheritdoc" */
  public function acceptThis5($p1, $p2, $p3, $p4, $x) {
    return $x->getThis();
  }
}

$foo = new Foo();
$_ = $foo->getThis()->getThis();
$foo->acceptThis($foo);
`)
	test.RunAndMatch()
}

func TestMagicGetChaining(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Magic {
  /** @return Magic */
  public function __get($key) {
    return $this;
  }

  /** Method that does nothing */
  public function method() {}
}

$m = new Magic();
$m->method();
$m->foo->method();
$m->foo->bar->method();
`)
}

func TestIteratorForeach(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
interface Iterator extends Traversable {
  public function current();
  public function key();
  public function next();
  public function rewind();
  public function valid();
}

class SimpleXMLIterator implements Iterator {
  /** @return int */
  public function blah($name) { return 0; }

  /** @return SimpleXMLIterator */
  public function current() {}
  /** @return int */
  public function key() {}
  /** @return void */
  public function next() {}
  /** @return void */
  public function rewind() {}
  /** @return bool */
  public function valid() {}
}

function testForeach($xml_str) {
  $xml = new SimpleXMLIterator($xml_str);
  foreach ($xml as $node) {
    $_ = $node->blah('123');
  }
}
`)
}

func TestSimpleXMLElementForeach(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class SimpleXMLElement implements Traversable, ArrayAccess {
  /** @return SimpleXMLElement */
  private function __get($name) {}

  /** @return static[] */
  public function xpath ($path) {}

  /** @return SimpleXMLElement */
  public function offsetGet($i) {}
}

function testForeach($xml_str) {
  $xml = new SimpleXMLElement($xml_str);
  foreach ($xml as $node) {
    $_ = $node->xpath('blah');
  }
}
`)
}

func TestSimpleXMLElement(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class SimpleXMLElement {
  /** @return SimpleXMLElement */
  private function __get($name) {}
  /** @return static[] */
  public function xpath ($path) {}
}

class SimpleXMLIterator extends SimpleXMLElement {
  /** @return SimpleXMLIterator|null */
  public function current () {}
}

function simpleElement($xml_str) {
  $el = new SimpleXMLElement("<a></a>", 0);
  $iters = $el->xpath("/a");
  $_ = $iters[0]->foo;
  $_ = $el->foo;
  $_ = $el->foo->bar;
  $_ = $el->foo->bar->xpath("/a");
}

function simpleIterator($xml_str) {
  $el = new SimpleXMLIterator("<a></a>", 0);
  $iters = $el->xpath("/a");
  $root = $iters[0];
  $_ = $iters[0]->current();
  $_ = $root->current();
  $_ = $root->current()->foo;
}

function simpleIteratorReassign($xml_string) {
  $el = new SimpleXMLIterator("<a></a>", 0);
  $iter = $el->xpath("/a");
  $iter = $iter[0];
  $_ = $iter->current();
}
`)
}

func TestLateStaticBindingForeach(t *testing.T) {
	// This test comes from https://youtrack.jetbrains.com/issue/WI-28728.
	// Phpstorm currently reports `hello` call in `$item->getC()->hello()`
	// as undefined. NoVerify manages to resolve it properly.

	linttest.SimpleNegativeTest(t, `<?php
class A
{
    /**
     * @return static[]
     */
    public function create()
    {
        return [new static()];
    }
}

class B extends A
{
    /**
     * @return C
     */
    public function getC()
    {
        return new C();
    }
}

class C
{
    /** Hello does nothing */
    public function hello()
    {
        return 'Hello world!';
    }
}

$b = new B();
foreach ($b->create() as $item) {
    $item->getC()->hello();
}
`)
}

func TestDerivedLateStaticBinding(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Base {
  /** @return static[] */
  public function asArray() { return []; }
}

class Derived extends Base {
  /**
    * Will cause "undefined method" warning if called via instance that
    * is returned by Base.asArray without late static binding support from NoVerify.
    */
  public function onlyInDerived() { return 1; }
}

$x = new Derived();
$xs = $x->asArray();
$_ = $xs[0]->onlyInDerived();
`)
}

func TestStaticResolutionInsideSameClass(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class NotA {
		/** @return static */
		public static function instance() {
			return new static;
		}
	}

	class A {
		/** @return int */
		public function b() {
			$a = new static();
			return $a->c();
		}
		protected function fakeB() {
			$a = NotA::instance();
			return $a->c();
		}
		protected function instance() {
			return new static;
		}
		/** @return int */
		public function c() {
			return 1;
		}
	}

	class B extends A {
		protected function derived() {}
		protected function test() {
			$this->instance()->derived();
			$this->instance()->nonDerived();
		}
	}`)

	test.Expect = []string{
		"Call to undefined method {\\NotA}->c()",
		"Call to undefined method {\\B}->nonDerived()",
	}
	test.RunAndMatch()
}

func TestInheritanceLoop(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class A extends B { }
	class B extends A { }

	function test() {
		return A::SOMETHING;
	}`)
	test.Expect = []string{"Class constant \\A::SOMETHING does not exist"}
	test.RunAndMatch()
}

func TestClosureLateBinding(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class Example
	{
		public function method()
		{
			return 42;
		}
	}

	class Closure {
		public function call();
	}

	(function() {
		$this->method();
		$a->method();
	})->call(new Example());
	`)
	test.Expect = []string{
		"Undefined variable: a",
		"Call to undefined method {mixed}->method()",
	}
	runFilterMatch(test, "undefined")
}

func TestProtected(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class A {
		private $priv;

		protected $prop;
		protected $prop2;
		protected static $static_prop;
		protected static $static_prop2;
		protected const C = 1;
		protected const C2 = 1;
		protected static function staticMethod() { }
		protected static function staticMethod2() { }
		protected function method() { }
		protected function methodFromClosure() { }
		protected function method2() { }
		protected function methodFromClosure2() { }
	}

	class B extends A {
		private $privB;

		public function okContext() {
			echo $this->priv;
			echo $this->privB;
			echo $this->prop;
			echo self::$static_prop;
			echo self::C;
			echo self::staticMethod();
			echo $this->method();

			(function() { echo $this->methodFromClosure(); })();
		}
	}

	class D {

	}

	class C extends D {
		public function wrongContext() {
			$instance = new B;

			echo $instance->prop2;
			echo B::$static_prop2;
			echo B::C2;
			echo B::staticMethod2();
			echo $instance->method2();

			(function() use($instance) { echo $instance->methodFromClosure2(); })();
		}
	}`)
	test.Expect = []string{
		`Cannot access private property \A->priv`,
		`Cannot access protected property \A->prop2`,
		`Cannot access protected property \A::$static_prop2`,
		`Cannot access protected constant \A::C2`,
		`Cannot access protected method \A->method2()`,
		`Cannot access protected method \A->methodFromClosure2()`,
		`Cannot access protected method \A::staticMethod2()`,
	}
	runFilterMatch(test, "accessLevel")
}

func TestInvoke(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class Example
	{
		public function __invoke($argument)
		{
			return 42;
		}
	}

	(new Example())();
	`)
	test.Expect = []string{"Too few arguments"}
	runFilterMatch(test, "argCount")
}

func TestTraversable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	interface Traversable {}
	interface Iterator extends Traversable {}
	interface IteratorAggregate extends Traversable {}
	interface SeekableIterator extends Iterator, SeekableIterator {} // test for loop detection
	class ArrayIterator implements SeekableIterator {}

	class Example implements \IteratorAggregate
	{
		public function someFunc()
		{
			return 1;
		}

		public function getIterator()
		{
			return [1, 2, 3];
		}
	}

	class ExampleGood implements \IteratorAggregate
	{
		public function someFunc() {
			return 1;
		}

		public function getIterator()
		{
			return new ArrayIterator([1, 2, 3]);
		}
	}

	foreach (new Example() as $i) {
	   echo $i;
	}`)
	test.Expect = []string{
		`Objects returned by \Example::getIterator() must be traversable or implement interface \Iterator`,
	}
	runFilterMatch(test, "stdInterface")
}

func TestInstanceOfElseif2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function fn3($f3) {
  if ($f3 instanceof File) {
    return $f3->name();
  } else if ($f3 instanceof Video) {
    return $f3->filename();
  }
  return "";
}

function fn4($f4) {
  if ($f4 instanceof File) {
    return $f4->name();
  } elseif ($f4 instanceof Video) {
    return $f4->filename();
  }
  return "";
}`)
	test.Expect = []string{
		`Call to undefined method {\File}->name()`,
		`Call to undefined method {\File|\Video}->filename()`,
		`Call to undefined method {\File}->name()`,
		`Call to undefined method {\File|\Video}->filename()`,
	}
	test.RunAndMatch()
}

func TestInstanceOfElseif1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class File {
  /** @return string */
  public function filename() { return ""; }
}
class Video {
  /** @return string */
  public function name() { return ""; }
}

function fn1($f1) {
  if ($f1 instanceof File) {
    return $f1->filename();
  } elseif ($f1 instanceof Video) {
    return $f1->name();
  }
  return "";
}

function fn2($f2) {
  if ($f2 instanceof File) {
    return $f2->filename();
  } else if ($f2 instanceof Video) {
    return $f2->name();
  }
  return "";
}
`)
}

func TestInstanceOf(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
	class Element {
		public function get(): int {
			return 0;
		}

		public function get2(): int {
			return 0;
		}
	}

	function doSomething($param) {}

	class Test {
		private $arr = [];

		public function simple($obj) {
			if ($obj instanceof Element) {
				return $obj->get();
			}
		}

		public function complexProp($key) {
			if (isset($this->arr[$key]) && $this->arr[$key] instanceof Element) {
				return $this->arr[$key]->get();
			}
		}

		public function complexCall($key) {
			if (doSomething($key) instanceof Element) {
				return doSomething($key)->get();
			}
		}

		public function invalidComplexCall($key) {
			if (doSomething($key) instanceof Element) {
				return doSomething($key . 'test')->get2();
			}
		}

		public function invalid($obj) {
			if ($obj instanceof Element) {
				return $obj->callUndefinedMethod();
			}
		}
	}`)
	test.Expect = []string{
		`Call to undefined method {void}->get2()`,
		`Call to undefined method {\Element}->callUndefinedMethod()`,
	}
	runFilterMatch(test, "undefined")
}


func TestNullableTypes(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
	/** @var ?B */
	public $b;
}

class B {
	public $c;
}

function test() {
	return (new A)->b->c;
}

function test2(?A $instance) {
	return $instance->b;
}

function test3(?A $instance) {
	return $instance->c;
}
`)
	test.Expect = []string{
		`Property {\A|null}->c does not exist`,
	}
	test.RunAndMatch()
}
