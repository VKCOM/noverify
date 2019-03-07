package linter

import (
	"log"
	"testing"
)

func TestInterfaceConstants(t *testing.T) {
	reports := getReportsSimple(t, `<?php

	interface TestInterface
	{
		const TEST = '1';
	}

	class TestClass implements TestInterface
	{
		public function get()
		{
			return self::TEST;
		}
	}
	`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestInheritanceLoop(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class A extends B { }
	class B extends A { }

	function test() {
		return A::SOMETHING;
	}
	`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, "Class constant \\A::SOMETHING does not exist") {
		t.Errorf("Class contant SOMETHING must be missing")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestReturnTypes(t *testing.T) {
	reports := getReportsSimple(t, `<?php
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
	}
	`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestVariadic(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class TestClass
	{
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

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestTraitProperties(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	declare(strict_types=1);

	trait Example
	{
		private static $property = 'some';

		protected function some(): string
		{
			return self::$property;
		}
	}
	`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestMagicMethods(t *testing.T) {
	reports := getReportsSimple(t, `<?php
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
	}
	`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestGenerator(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class Generator {
		public function send();
	}

	function a($a): \Generator
	{
		yield $a;
	}

	a(42)->send(42);
	`)

	if len(reports) != 0 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestClosureLateBinding(t *testing.T) {
	reports := getReportsSimple(t, `<?php
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

	if len(reports) != 2 {
		t.Errorf("Unexpected number of reports: expected 2, got %d", len(reports))
	}

	if !hasReport(reports, "Undefined variable: a") {
		t.Errorf("Must be a warning about undefined variable a")
	}

	if !hasReport(reports, "Call to undefined method {}->method()") {
		t.Errorf("Must be an error about call to undefined method()")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestInterfaceInheritance(t *testing.T) {
	reports := getReportsSimple(t, `<?php
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
	}
	`)

	if len(reports) != 3 {
		t.Errorf("Unexpected number of reports: expected 3, got %d", len(reports))
	}

	if !hasReport(reports, `Call to undefined method {\TestExInterface}->nonexistent()`) {
		t.Errorf("Must be an error about call to nonexistent")
	}

	if !hasReport(reports, "Call to undefined method {}->format()") {
		t.Errorf("Must be an error about call to format of undefined")
	}

	if !hasReport(reports, "Class constant \\TestExInterface::TEST2 does not exist") {
		t.Errorf("Must be an error about missing class constant TEST2")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestProtected(t *testing.T) {
	reports := getReportsSimple(t, `<?php
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
	}
	`)

	if len(reports) != 7 {
		t.Errorf("Unexpected number of reports: expected 5, got %d", len(reports))
	}

	if !hasReport(reports, `Cannot access private property \A->priv`) {
		t.Errorf("Must be an error about access to private property")
	}

	if !hasReport(reports, `Cannot access protected property \A->prop2`) {
		t.Errorf("Must be an error about access to property")
	}

	if !hasReport(reports, `Cannot access protected property \A::$static_prop2`) {
		t.Errorf("Must be an error about access to static property")
	}

	if !hasReport(reports, `Cannot access protected constant \A::C2`) {
		t.Errorf("Must be an error about access to constant")
	}

	if !hasReport(reports, `Cannot access protected method \A->method2()`) {
		t.Errorf("Must be an error about call to method")
	}

	if !hasReport(reports, `Cannot access protected method \A->methodFromClosure2()`) {
		t.Errorf("Must be an error about call to method from inside a closure")
	}

	if !hasReport(reports, `Cannot access protected method \A::staticMethod2()`) {
		t.Errorf("Must be an error about call to static method")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestInvoke(t *testing.T) {
	reports := getReportsSimple(t, `<?php
	class Example
	{
		public function __invoke($argument)
		{
			return 42;
		}
	}

	(new Example())();
	`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, `Too few arguments`) {
		t.Errorf("Must be an error about too few arguments to expression")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestTraversable(t *testing.T) {
	reports := getReportsSimple(t, `<?php
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
	}
	`)

	if len(reports) != 1 {
		t.Errorf("Unexpected number of reports: expected 1, got %d", len(reports))
	}

	if !hasReport(reports, `Objects returned by \Example::getIterator() must be traversable or implement interface \Iterator`) {
		t.Errorf("Must be an error about getIterator()")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}

func TestInstanceOf(t *testing.T) {
	reports := getReportsSimple(t, `<?php
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
	}
	`)

	if len(reports) != 2 {
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
	}

	if !hasReport(reports, `Call to undefined method {}->get2()`) {
		t.Errorf("Must be an error in invalidComplexCall()")
	}

	if !hasReport(reports, `Call to undefined method {\Element}->callUndefinedMethod()`) {
		t.Errorf("Must be an error in invalid()")
	}

	for _, r := range reports {
		log.Printf("%s", r)
	}
}
