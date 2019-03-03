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

	if !hasReport(reports, "Class constant does not exist") {
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
		t.Errorf("Unexpected number of reports: expected 0, got %d", len(reports))
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
