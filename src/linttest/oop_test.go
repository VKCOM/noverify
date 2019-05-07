package linttest_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

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
    public static function create()
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

func TestLateStaticBinding(t *testing.T) {
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
  public function onlyInDerived() {}
}

$x = new Derived();
$xs = $x->asArray();
$_ = $xs[0]->onlyInDerived();
`)
}

func TestInterfaceConstants(t *testing.T) {
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

func TestReturnTypes(t *testing.T) {
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

func TestVariadic(t *testing.T) {
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
	`)
}

func TestTraitProperties(t *testing.T) {
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

func TestMagicMethods(t *testing.T) {
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

func TestGenerator(t *testing.T) {
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
		"Call to undefined method {}->method()",
	}
	runFilterMatch(test, "undefined")
}

func TestInterfaceInheritance(t *testing.T) {
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
		"Call to undefined method {}->format()",
		"Class constant \\TestExInterface::TEST2 does not exist",
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
		`Call to undefined method {}->get2()`,
		`Call to undefined method {\Element}->callUndefinedMethod()`,
	}
	runFilterMatch(test, "undefined")
}
