package linttest_test

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
		`Cannot instantiate abstract class`,
		`Cannot instantiate abstract class`,
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
interface Traversable {}

interface ArrayAccess {
  public function offsetExists($offset);
  public function offsetGet($offset);
  public function offsetSet($offset, $value);
  public function offsetUnset($offset);
}

class SimpleXMLElement implements Traversable, ArrayAccess {
  /** @return SimpleXMLElement */
  private function __get($name) {}

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
	linttest.RunFilterMatch(test, "undefined")
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
	test.Expect = []string{"Too few arguments"}
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
	linttest.RunFilterMatch(test, "undefined")
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
		`Type \T7\UnknownClass not found`,
		`Type \T8\UnknownTrait not found`,
		`Type \T6\UnknownIface not found`,

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
	linttest.RunFilterMatch(test, `unimplemented`, `nameCase`, `undefined`)
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
	test.RunAndMatch()
}
