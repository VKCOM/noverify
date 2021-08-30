package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestParentConstructorCall(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Leaf {
  public function __construct() {}
}

class WithDefaultCtor {}

// OK: WithDefaultCtor doesn't have its own constructor.
class Foo extends WithDefaultCtor {
  public function __construct() {}
}

// OK: Bar does not define its own constructor.
class Bar extends Foo {}

class WithTwoParams {
  public function __construct($x, $y) {}
}

class WithAbstract {
  abstract public function __construct() {}
}
`)
	test.AddFile(`<?php
class Good1 extends Foo {
  public function __construct() {
    parent::__construct();
  }
}

class Good2 extends WithTwoParams {
  public function __construct() {
    echo 123;
    {
      parent::__construct(1, 2);
    }
  }
}

class Good3 extends WithAbstract {
  public function __construct() {}
}
`)
	test.AddFile(`<?php
class Bad1 extends Foo {
  public function __construct() {
    echo 123;
  }
}

class Bad2 extends WithTwoParams {
  public function __construct() {
    echo 123;
    {
      return;
    }
  }
}
`)
	test.Expect = []string{
		`Missing parent::__construct()`,
		`Missing parent::__construct()`,
	}
	test.RunAndMatch()
}

func TestParentConstructorCallWithOtherStaticCall(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {
	public function __construct(int $a) {}
}

class Boo extends Foo {
  /** doc */
  public static function getId(): int {}

  public function __construct() {
    parent::__construct(self::getId());
  }
}

class Goo extends Foo {
  /** doc */
  public static function getId(): int {}

  public function __construct() {
    parent::__construct(100);
    $_ = self::getId();
  }
}
`)
}

func TestNewAbstract(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
abstract class AC {
  private static function foo() {
    return new self(); // Same as AC
  }

  /** @return static */
  public static function bar() {
    return new static(); // OK: late static binding
  }
}

$x = new AC();
`)
	test.Expect = []string{
		`Cannot instantiate abstract class \AC`,
		`Cannot instantiate abstract class \AC`,
	}
	test.RunAndMatch()
}

func TestNewInterface(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
interface IElement {}
$x = new IElement;
`)
	test.Expect = []string{
		`Cannot instantiate interface \IElement`,
	}
	test.RunAndMatch()
}

func TestNewTrait(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait Element {}
$x = new Element;
`)
	test.Expect = []string{
		`Cannot instantiate trait \Element`,
	}
	test.RunAndMatch()
}

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

trait TraitWithNameMatchingMethod {
  /** ok */
  public function TraitWithNameMatchingMethod() {}
}

interface InterfaceWithNameMatchingMethod {
  /** ok */
  public function InterfaceWithNameMatchingMethod();
}

namespace SameWithNamespace {
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

  trait TraitWithNameMatchingMethod {
    /** ok */
    public function TraitWithNameMatchingMethod() {}
  }
  
  interface InterfaceWithNameMatchingMethod {
    /** ok */
    public function InterfaceWithNameMatchingMethod();
  }
}
`)
	test.Expect = []string{
		`Old-style constructor usage, use __construct instead`,
		`Old-style constructor usage, use __construct instead`,
		`Old-style constructor usage, use __construct instead`,
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
		`Too few arguments for \OneArg constructor, expecting 1, saw 0`,
		`Too few arguments for \A\B\TwoArgs constructor, expecting 2, saw 0`,
		`Too few arguments for \A\B\TwoArgs constructor, expecting 2, saw 0`,
		`Too few arguments for \OneArgDerived constructor, expecting 1, saw 0`,
	}
	test.RunAndMatch()
}

func TestPHPDocProperty(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * @property int $int
 * @property string $string - optional description.
 * @property-read string $name Name of the class
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

/**
 * @param WithProps|null $y
 */
function f2($y = null) {
  echo $y->name;
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

func TestPHPDocPropertyForClassWithModifiers(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * @property int $int
 * @property string $string - optional description.
 * @property-read string $name Name of the class
 */
abstract class WithPropsBase {
  /***/
  public function getInt() {
    return $this->int;
  }
  /***/
  public function getString() {
    return $this->string;
  }
}

/**
 * @property int $int1
 * @property string $string1 - optional description.
 * @property-read string $name1 Name of the class
 */
final class WithProps extends WithPropsBase {}

$_ = (new WithProps())->int;
$_ = (new WithProps())->string;

function f(WithProps $x) {
  return $x->int + $x->int1;
}

/**
 * @param WithProps|null $y
 */
function f2($y = null) {
  echo $y->name;
  echo $y->string1;
}

// Can't access them as static props/constants.
$_ = WithProps::int;
$_ = WithProps::int1;
$_ = WithProps::$int;
$_ = WithProps::$int1;
`)

	test.Expect = []string{
		`Class constant \WithProps::int does not exist`,
		`Class constant \WithProps::int1 does not exist`,
		`Property \WithProps::$int does not exist`,
		`Property \WithProps::$int1 does not exist`,
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

func TestNonPublicMagicMethods(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
  public static function __set($name, $value) {} // The magic method __set() cannot be static
  protected function __toString() {} // The magic method __call() must have public visibility
  public function __callStatic($name, $arguments) {} // The magic method __callStatic() must be static
  public static function __destruct() {} // The magic method __destruct() cannot be static
  public static function __construct($name, $arguments) {} // The magic method __construct() cannot be static
  private static function __call($name, $arguments) {} // The magic method __call() must have public visibility
  private static function __callStatic($name, $arguments) {} // The magic method __callStatic() must have public visibility
}`)

	test.Expect = []string{
		"The magic method __set() cannot be static",
		"The magic method __toString() must have public visibility",
		"The magic method __callStatic() must be static",
		"The magic method __destruct() cannot be static",
		"The magic method __construct() cannot be static",
		"The magic method __call() cannot be static",
		"The magic method __call() must have public visibility",
		"The magic method __callStatic() must have public visibility",
	}
	test.RunAndMatch()
}

func TestNonPublicMagicMethodsGood(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class A {
  public function __call($name, $arguments) {} // Ok
  public function __toString() {} // Ok
  public function __set($name, $value) {} // Ok
  public function __get($name) {} // Ok
  public function __clone() {} // Ok

  public function __set_state() {} // Ok
  private function __set_state() {} // Ok, can be non-public
  protected function __set_state() {} // Ok, can be non-public
  public static function __set_state() {} // Ok, can be static

  public function __sleep() {} // Ok
  private function __sleep() {} // Ok, can be non-public
  protected function __sleep() {} // Ok, can be non-public
  public static function __sleep() {} // Ok, can be static

  public function __wakeup() {} // Ok
  private function __wakeup() {} // Ok, can be non-public
  protected function __wakeup() {} // Ok, can be non-public
  public static function __wakeup() {} // Ok, can be static

  public function __serialize() {} // Ok
  private function __serialize() {} // Ok, can be non-public
  protected function __serialize() {} // Ok, can be non-public
  public static function __serialize() {} // Ok, can be static

  public function __unserialize() {} // Ok
  private function __unserialize() {} // Ok, can be non-public
  protected function __unserialize() {} // Ok, can be non-public
  public static function __unserialize() {} // Ok, can be static

  public function __clone() {} // Ok
  private function __clone() {} // Ok, can be non-public
  protected function __clone() {} // Ok, can be non-public

  public function __construct() {} // Ok
  private function __construct() {} // Ok
  protected function __construct() {} // Ok
  public static function __callStatic($name, $arguments) {} // Ok
  private static function __some_method() {} // Ok, not magic
}`)
}

func TestWrongNumberOfArgumentsInMagicMethods(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {
  public function __destruct($a) {} // 0
  public function __call($a) {} // 2
  public static function __callStatic($a, $b, $c) {} // 2
  public function __get($a, $b) {} // 1
  public function __set($a) {} // 2
  public function __isset() {} // 1
  public function __unset($a, $b) {} // 1
  public function __toString($a) {} // 0
}`)

	test.Expect = []string{
		"The magic method __destruct() must take exactly 0 argument",
		"The magic method __call() must take exactly 2 argument",
		"The magic method __callStatic() must take exactly 2 argument",
		"The magic method __get() must take exactly 1 argument",
		"The magic method __set() must take exactly 2 argument",
		"The magic method __isset() must take exactly 1 argument",
		"The magic method __unset() must take exactly 1 argument",
		"The magic method __toString() must take exactly 0 argument",
	}
	test.RunAndMatch()
}

func TestWrongNumberOfArgumentsInMagicMethodsGood(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {
  public function __construct($a) {}
  public function __construct($a, $b) {}
  public function __destruct() {} // 0
  public function __call($a, $b) {} // 2
  public static function __callStatic($a, $b) {} // 2
  public function __get($a) {} // 1
  public function __set($a, $b) {} // 2
  public function __isset($a) {} // 1
  public function __unset($a) {} // 1
  public function __sleep($a) {}
  public function __wakeup() {}
  public function __serialize() {}
  public function __unserialize() {}
  public function __toString() {} // 0
  public function __invoke() {}
  public function __set_state() {}
  public function __clone() {} // 0
  public function __debugInfo() {}
  public function __non_magic($a) {} // any
  public function __non_magic() {} // any
}`)
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
interface Traversable {}

interface ArrayAccess {
  public function offsetExists($offset);
  public function offsetGet($offset);
  public function offsetSet($offset, $value);
  public function offsetUnset($offset);
}

class SimpleXMLElement implements Traversable, ArrayAccess {
  /** @return SimpleXMLElement */
  public function __get($name) {}

  /** @return static[] */
  public function xpath ($path) {}

  /** @return SimpleXMLElement */
  public function offsetGet($i) {}

  /** @inheritdoc */
  public function offsetExists($offset) {}
  /** @inheritdoc */
  public function offsetSet($offset, $value) {}
  /** @inheritdoc */
  public function offsetUnset($offset) {}
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
  public function __get($name) {}
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
    echo $item->getC()->hello();
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

func TestStaticResolutionInsideOtherStaticResolution(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
	class SomeMainClass extends ParentClass {
		/**
		* @var string
		*/
		public $testProperty = '';
		
		/**
		* Some test method
		*/
		public function testMethod() {}
	}
	
	class ParentClass {
		/** @return static Some */
		public static function findOne() {
		// static here
			return static::findByCondition();
		}
		
		/** @return static Some */
		public static function findByCondition() {
			// and static here
			$_ = static::find();
			return new static();
		}
		
		/** @return int object_id */
		public static function find() {
			return 0;
		}
	}
	
	function f() {
		$result = '';
		
		$objectB = SomeMainClass::findOne();
		$objectB->testMethod();
		
		$result .= $objectB->testProperty;
		echo $result;
	}`)
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
	test.Config().StrictMixed = true
	test.AddFile(`<?php
	class Example
	{
		public function method()
		{
			return 42;
		}
	}

	(function() {
		$this->method();
		$a->method();
	})();
	`)
	test.Expect = []string{
		"Cannot find referenced variable $a",
		"Call to undefined method {undefined}->method()",
	}
	linttest.RunFilterMatch(test, "undefinedVariable", "undefinedMethod")
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
	linttest.RunFilterMatch(test, "accessLevel")
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
	test.Expect = []string{"Too few arguments for __invoke, expecting 1, saw 0"}
	linttest.RunFilterMatch(test, "argCount")
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
	linttest.RunFilterMatch(test, "stdInterface")
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
		`Call to undefined method {\Video}->filename()`,
		`Call to undefined method {\File}->name()`,
		`Call to undefined method {\Video}->filename()`,
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
		`Call to undefined method {\Element}->callUndefinedMethod()`,
	}
	linttest.RunFilterMatch(test, "undefinedMethod")
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

func TestAbstractClassMethodCall(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
interface FooInterface {
  public function foo();
}

abstract class AbstractBarBase implements FooInterface {
  protected function callIface() {
    $this->foo(); // OK: abstract class can call unimplemented iface methods
  }

  abstract protected function bar() : int;

  private function callBar() {
    return $this->bar(); // OK: can call to own abstract methods
  }
}

abstract class AbstractBar extends AbstractBarBase {
  protected function callIface2() {
    $this->foo(); // OK: as with AbstractBarBase which we extend
  }
  protected function callBar() {
    return $this->bar(); // OK: can call base class abstract methods
  }
}

abstract class AbstractBar2 extends AbstractBar {
  /** Implements FooInterface */
  public function foo() {}
}

class BarImpl extends AbstractBar2 {
  protected function callIface3() {
    $this->foo(); // OK: calling interface method implemented in a base class
  }
  protected function callBar2() {
    return $this->bar(); // OK: calling own method that implements abstract method
  }

  protected function bar() : int {
    return 10;
  }
}

function f(AbstractBarBase $x) {
  return $x->foo();
}
`)
}

func TestUnimplemented(t *testing.T) {
	test := linttest.NewSuite(t)

	test.AddFile(`<?php
namespace T1;

interface AB {
  public function a() {}
  public function b() {}
}


// Doesn't implement b(), but it's OK for abstract classes.
abstract class ImplementsA implements AB {
  public function a() {}
  abstract protected function fromAbstract();
}

class ImplementsAB extends ImplementsA {
  public function b() {}
  protected function fromAbstract() {}
}

class Bad extends ImplementsA implements AB {}
`)

	test.AddFile(`<?php
namespace T2;

// T2 differs from T1 by a fact that ImplementsAB and Bad do not
// specify "implements AB" directly.

interface AB {
  public function a() {}
  public function b() {}
}

// Doesn't implement b(), but it's OK for abstract classes.
abstract class ImplementsA implements AB {
  public function a() {}
  abstract protected function fromAbstract();
}

class ImplementsAB extends ImplementsA {
  public function b() {}
  protected function fromAbstract() {}
}

class Bad extends ImplementsA implements AB {}
`)

	test.AddFile(`<?php
namespace T3;

// Test interface inheritance and traits.

interface IfaceA { public function a(); }
interface IfaceB { public function b(); }
interface IfaceAB extends IfaceA, IfaceB {}

trait TraitA { public function a() {} }
trait TraitB { public function b() {} }
trait TraitAB { use TraitA; use TraitB; }

class Bad1 implements IfaceA, IfaceB {}
class Bad2 implements IfaceAB {}

class Good1 implements IfaceAB {
  use TraitA;
  use TraitB;
}

class Good2 implements IfaceAB {
  use TraitAB;
}

class AlmostGood implements \t1\AB {
  use TraitAB;
}

class Good3 extends Good2 {}
class Good4 extends Good3 implements IfaceAB {}
`)

	test.AddFile(`<?php
namespace T4;

// Test that abstract class inheritance chain is checked.

abstract class AbstractA {
  abstract protected function a();
}
abstract class AbstractAB extends AbstractA {
  abstract protected function b();
}

class Bad extends AbstractAB {}
`)

	test.AddFile(`<?php
namespace T5;

// Test that case-mismatching names still implement an interface.

interface IfaceFoo {
  public function foo();
}

abstract class AbstractFoo {
  abstract public function Foo();
}

class BadCase1 implements IfaceFoo {
  public function Foo() {}
}

class BadCase2 extends AbstractFoo {
  public function foo() {}
}
`)

	test.AddFile(`<?php
namespace T6;

// Case that addresses a case found in Yii2.
// Level1 trait defines abstract method.
// Level2 (parent) class implements that abstract method.

trait TraitAbstractA {
  abstract protected function a();
}

class BaseClass implements UnknownIface {
  protected function a();
}

class C extends BaseClass {
  use TraitAbstractA;
}

class Bad {
  use TraitAbstractA;
}
`)

	test.AddFile(`<?php
namespace T7;

// Like T6, but also with interface involved.
// Note that BaseClass is not declared as implementing IfaceA.

interface IfaceA {
  public function a();
}

trait TraitAbstractA {
  abstract public function a();
}

class BaseClass extends UnknownClass {
  public function a() {}
}

class C1 extends BaseClass implements Ifacea {
  use TraitAbstractA;
}

// Same as C1, but without a trait.
class C2 extends BaseClass implements IfaceA {}
`)

	test.AddFile(`<?php
namespace T8;

// Like abstract classes, trait can leave a contract unimplemented.

trait AbstractTraitA {
  abstract public function a();
}
trait AbstractTraitB {
  abstract public function b();
}

trait AbstractTraitAB {
  use AbstracttraitA;
  use AbstractTraitB;
  use UnknownTrait;
}
`)

	test.Expect = []string{
		`Class or interface named \T7\UnknownClass does not exist`,
		`Class or interface named \T6\UnknownIface does not exist`,
		`Trait named \T8\UnknownTrait does not exist`,

		`\t1\AB should be spelled \T1\AB`,
		`\T5\BadCase1::Foo should be spelled as \T5\IfaceFoo::foo`,
		`\T5\BadCase2::foo should be spelled as \T5\AbstractFoo::Foo`,
		`\T7\Ifacea should be spelled \T7\IfaceA`,

		`Class \T1\Bad must implement \T1\AB::b method`,
		`Class \T1\Bad must implement \T1\ImplementsA::fromAbstract`,

		`Class \T2\Bad must implement \T2\AB::b method`,
		`Class \T2\Bad must implement \T2\ImplementsA::fromAbstract method`,

		`Class \T3\Bad1 must implement \T3\IfaceA::a method`,
		`Class \T3\Bad1 must implement \T3\IfaceB::b method`,
		`Class \T3\Bad2 must implement \T3\IfaceA::a method`,
		`Class \T3\Bad2 must implement \T3\IfaceB::b method`,

		`Class \T4\Bad must implement \T4\AbstractAB::b method`,
		`Class \T4\Bad must implement \T4\AbstractA::a method`,

		`Class \T6\Bad must implement \T6\TraitAbstractA::a method`,
	}
	linttest.RunFilterMatch(test, "unimplemented", "nameMismatch", "undefinedClass", "undefinedTrait")
}

func TestInterfaceRules(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
interface WithConstants {
  const r = 10000; // ok
  public const v = 1; // ok
  private const b = 100, c1 = 1000; // 'b' can't be private, 'c1' can't be private
  protected const c = 100; // 'c' can't be protected
}

interface WithMethods {
  public function c(); // ok
  private function b(); // 'b' can't be private
  protected function f(); // 'f' can't be protected
}

interface WithStaticMethods {
  static function f1(); // ok
  public static function f2(); // ok
  private static function bad1(); // 'bad1' can't be private
  static protected function bad2(); // 'bad2' can't be protected
}

interface WithoutAnyModifier {
    function f(); // Ok,
}

`)
	test.Expect = []string{
		`'b' can't be private`,
		`'c1' can't be private`,
		`'c' can't be protected`,
		`'b' can't be private`,
		`'f' can't be protected`,
		`'bad1' can't be private`,
		`'bad2' can't be protected`,
	}
	linttest.RunFilterMatch(test, "nonPublicInterfaceMember")
}

func TestMixinAnnotation(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
namespace QQ\WW;

/**
 * @mixin SomeQQClass
 */
class SomeQQClass1 {
  /** */
  public function methodQQ1()
  {
    echo "";
  }
}

/**
 * @mixin SomeQQClass1
 */
class SomeQQClass {
  /** */
  public function methodQQ()
  {
    echo "";
  }
}
`)

	test.AddFile(`<?php
use QQ\WW\SomeQQClass;
use QQ\WW\SomeQQClass1 as SomeQQClass2;

/**
 * @mixin SomeQQClass
 */
class SomeClass {
  /** */
  public function method()
  {
    echo $this->methodQQ1();
  }
}

/**
 * @mixin SomeClass
 */
class SomeClass2 {
  /** */
  public function method2()
  {
    echo $this->method(); // Ok, from mixin SomeClass
  }
}

/** 
 * @mixin SomeQQClass2
 */
class SomeClass3 {
  /** */
  public function method3()
  {
    echo "";
  }
}

/**
 * @mixin \SomeClass2
 * @mixin 
 */
class BarWithSomeMixin {
  /** */
  public function run()
  {
    $this->method(); // Ok, from mixin SomeClass
    $this->method2(); // Ok, from mixin SomeClass2
    $this->method3(); // Error, no SomeClass3 mixin
  }

  /** */
  public function barWithSomeMixinMethod()
  {
    echo "";
  }
}

/**
 * @mixin \BarWithSomeMixin
 * @mixin SomeClass3
 * @mixin Boo // Error, Boo not found
 */
class Bar {
  /** */
  public function run()
  {
    $this->method(); // Ok, from mixin SomeClass
    $this->method2(); // Ok, from mixin SomeClass2
    $this->method3(); // Ok, from mixin SomeClass3
    $this->barWithSomeMixinMethod(); // Ok, from mixin BarWithSomeMixin
  }
}
`)
	test.Expect = []string{
		`Call to undefined method {\BarWithSomeMixin}->method3()`,
		`@mixin tag refers to unknown class \Boo`,
	}
	test.RunAndMatch()
}

func TestGroupUse(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
namespace Test;

class TestClass {}
class TestClass2 {}

function testFunction() {}
function testFunction2() {}
`)

	test.AddFile(`<?php
namespace Test\Something;

class TestSomethingClass {}
class TestSomethingClass2 {}

function testSomethingFunction() {}
function testSomethingFunction2() {}
`)

	test.AddFile(`<?php
namespace Foo;

use Test\{
	TestClass, 
	TestClass2 as SomeClass,

	Something\TestSomethingClass, 
	Something\TestSomethingClass2 as SomethingClass
};

use function Test\{
	testFunction, 
	testFunction2 as someFunc,

	Something\testSomethingFunction,
	Something\testSomethingFunction2 as somethingFunc
};

function f() {
    $_ = new TestClass();
    $_ = new SomeClass();
    $_ = new TestSomethingClass();
    $_ = new SomethingClass();

    $_ = testFunction();
    $_ = someFunc();
    $_ = testSomethingFunction();
    $_ = somethingFunc();
}
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestTypeHintClassCaseFunctionParam(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {}

class Boo {
	/** */
	public function a2(foo $b) {}
}

function a1(Foo $a, foo $b, boo $c) {}
`)

	test.Expect = []string{
		`\foo should be spelled \Foo`,
		`\foo should be spelled \Foo`,
		`\boo should be spelled \Boo`,
	}
	test.RunAndMatch()
}

func TestClassComponentsOrder(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {
  public function f() {}

  const A = 10;
  public string $a = "";
}

class Foo1 {
  const A = 10;

  public function f() {}

  public string $a = "";
}

class Foo2 {
  const A = 10;

  public function f() {}

  public static string $a = "";

  public function f1() {}
}

class Foo3 {
  public function f() {}

  public string $a = "";

  public static function f1() {}

  const A = 10;
}
`)

	test.Expect = []string{
		`Constant A must go before methods in the class Foo`,
		`Property $a must go before methods in the class Foo`,
		`Property $a must go before methods in the class Foo1`,
		`Property $a must go before methods in the class Foo2`,
		`Property $a must go before methods in the class Foo3`,
		`Constant A must go before methods in the class Foo3`,
	}
	linttest.RunFilterMatch(test, "classMembersOrder")
}

func TestClassComponentsOrderGood(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Foo {
  const A = 10;
  public string $a = "";

  /** */
  public function f() {}
}

class Foo1 {
  const A = 10;
  public string $a = "";
}

class Foo2 {}

class Foo3 {
  const A = 10;

  /** */
  public function f() {}
}

class Foo4 {
  public string $a = "";

  /** */
  public function f() {}
}

class Foo5 {
  const A = 10, B = 100;
  public string $a = "", $b = "1";
  public string $c = "", $d = "1";

  /** */
  public function f() {}

  /** */
  public function f1() {}
}
`)
}

func TestCallStaticWithVariable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {
  /** */
  public static function some_method() {}
}
function ret_int() {
  return 12;
}
function ret_string() {
  return "Foo";
}
function ret_object() {
  return new Foo();
}
function f($arg) {
  $foo = new Foo();
  $foo::some_method(); // Ok
  $foo::non_existing_method(); // Error

  $a = 10;
  $a::some_method(); // invalid class name

  $foo2 = new Foo();
  $foo3 = new Foo();
  $foo4 = $arg;

  $foo5 = ret_string();
  $foo6 = ret_object();
  $foo7 = ret_int();

  if ($a > 100) {
    $foo2 = "Foo";
    $foo3 = 10;
  }

  $foo2::some_method(); // Skip, via \Foo|string type (both is correct for class name)
  $foo3::some_method(); // Error, int type is invalid class name
  $foo4::some_method(); // Skip, via mixed type
  $foo5::some_method(); // Ok ret_string returns the string
  $foo6::some_method(); // Ok ret_object returns the Foo object
  $foo7::some_method(); // Error, ret_int returns the int
}
`)
	test.Expect = []string{
		`Call to undefined method \Foo::non_existing_method()`,
	}
	test.RunAndMatch()
}

func TestImplicitAccessModifiers(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Foo {
  const FOO = 100; // ok
  public const FOO1 = 100; // ok
  private const FOO2 = 100; // ok
  protected const FOO3 = 100; // ok
  
  var int $prop = 100;
  public int $prop1 = 100;
  private int $prop2 = 100;
  protected int $prop3 = 100;

  var int $prop4, $prop5 = 100;
  public int $prop6, $prop7 = 100;
  private int $prop8, $prop9 = 100;
  protected int $prop10, $prop11 = 100;
  
  function f1() {}
  public function f2() {}
  private function f3() {}
  protected function f4() {}

  static function f5() {}
  public static function f6() {}
  private static function f7() {}
  protected static function f8() {}
}
`)
	test.Expect = []string{
		`Specify the access modifier for property explicitly`,
		`Specify the access modifier for properties explicitly`,
		`Specify the access modifier for \Foo::f1 method explicitly`,
		`Specify the access modifier for \Foo::f5 method explicitly`,
	}
	linttest.RunFilterMatch(test, "implicitModifiers")
}
