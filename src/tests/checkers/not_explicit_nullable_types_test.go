package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestNotNullableString(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function nullableString(?string $a = null) {}
`)

	test.RunAndMatch()
}

func TestNotNullableArray(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/** @param string[] array */
function nullableArray(array $a = null) {}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}
	test.RunAndMatch()
}

func TestNullableCallable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function nullableCallable(?callable $a = null) {}
`)

	test.RunAndMatch()
}

func TestNotNullableCallable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function NotNullableCallable(callable $a = null) {}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}
	test.RunAndMatch()
}

func TestNotNullableClasses(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class MyClass1 {}

class MyClass2 {
    public function myMethod(MyClass1 $a = null) {}
}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
		"Missing PHPDoc for",
	}
	test.RunAndMatch()
}

func TestNullableClasses(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class MyClass1 {}

class MyClass2 {
    public function myMethod(?MyClass1 $a = null) {}
}
`)

	test.Expect = []string{
		"Missing PHPDoc for",
	}
	test.RunAndMatch()
}

func TestNullableMultipleArgs(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function multipleArgsExample(?string $a, ?int $b = null, ?bool $c = null) {}
`)

	test.RunAndMatch()
}

func TestNotNullableMultipleArgs(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function multipleArgsExample(string $a, int $b = null, bool $c = null) {}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
		"parameter with null default value should be explicitly nullable",
	}

	test.RunAndMatch()
}

func TestNullableOrString(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function nullableOrString(null|string $a = null) {}
`)

	test.RunAndMatch()
}

func TestNullableOrClass(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class MyClass1 {}

class MyClass2 {
    public function myMethod(null|MyClass1 $a = null) {}
}
`)

	test.Expect = []string{
		"Missing PHPDoc for \\MyClass2::myMethod public method",
	}

	test.RunAndMatch()
}

func TestMixedParam(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function mixedParam(mixed $a = null) {}
`)

	test.RunAndMatch()
}

func TestTraitMethodNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait MyTrait {
    public function myMethod(?string $a = null) {}
}
`)

	test.Expect = []string{
		"Missing PHPDoc for \\MyTrait::myMethod public method",
	}

	test.RunAndMatch()
}

func TestTraitMethodNotNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait MyTrait {
    public function myMethod(string $a = null) {}
}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
		"Missing PHPDoc for \\MyTrait::myMethod public method",
	}

	test.RunAndMatch()
}

func TestInterfaceMethodNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
interface MyInterface {
    public function myMethod(?string $a = null);
}
`)

	test.RunAndMatch()
}

func TestInterfaceMethodNotNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
interface MyInterface {
    public function myMethod(string $a = null);
}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}

	test.RunAndMatch()
}

func TestClosureNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$closure = function(?string $a = null) {};
`)

	test.RunAndMatch()
}

func TestClosureNotNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$closure = function(string $a = null) {};
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}

	test.RunAndMatch()
}

func TestArrowFunctionNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$closure_fn = fn(?string $a = null) => null;
`)

	test.RunAndMatch()
}

func TestArrowFunctionNotNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$closure_fn = fn(string $a = null) => null;
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}

	test.RunAndMatch()
}

func TestArrowFunctionClassNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class MyClass {}

$closure_fn = fn(?MyClass $a = null) => null;
`)

	test.RunAndMatch()
}

func TestArrowFunctionClassNotNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class MyClass {}

$closure_fn = fn(MyClass $a = null) => null;
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}

	test.RunAndMatch()
}

func TestArrowFunctionArrayNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$closure_fn = fn(?array $a = null) => null;
`)

	test.RunAndMatch()
}

func TestArrowFunctionArrayNotNullable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$closure_fn = fn(array $a = null) => null;
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}

	test.RunAndMatch()
}

func TestArrowFunctionMultipleParams(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
$closure_fn = fn(?string $a, int $b = null, ?bool $c = null) => null;
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}

	test.RunAndMatch()
}
