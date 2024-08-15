package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestNotNullableString(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function nullableString(?string $a = null) {
	return 0;
}
`)

	test.RunAndMatch()
}

func TestNotNullableArray(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
* @param string[] array
*/
function nullableArray(array $a = null) {
	return 0;
}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}
	test.RunAndMatch()
}

func TestNullableCallable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function nullableCallable(?callable $a = null) {
    return 0;
}
`)

	test.RunAndMatch()
}

func TestNotNullableCallable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function NotNullableCallable(callable $a = null) {
    return 0;
}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
	}
	test.RunAndMatch()
}

func TestNotNullableClasses(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class MyClass1 {
}

class MyClass2 {
    public function myMethod(MyClass1 $a = null) {
        return 0;
    }
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
class MyClass1 {
}

class MyClass2 {
    public function myMethod(?MyClass1 $a = null) {
        return 0;
    }
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
function multipleArgsExample(?string $a, ?int $b = null, ?bool $c = null) {
	return 0;
}
`)

	test.RunAndMatch()
}

func TestNotNullableMultipleArgs(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function multipleArgsExample(string $a, int $b = null, bool $c = null) {
	return 0;
}
`)

	test.Expect = []string{
		"parameter with null default value should be explicitly nullable",
		"parameter with null default value should be explicitly nullable",
	}

	test.RunAndMatch()
}
