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
