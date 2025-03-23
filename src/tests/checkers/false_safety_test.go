package checkers

import (
	"github.com/VKCOM/noverify/src/linttest"
	"testing"
)

func TestFunctionPassingFalse_SimpleVar(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function test(string $s): void {
    echo $s;
}
$var = false;
test($var);
`)
	test.Expect = []string{
		"false passed to non-falseable parameter s in function test",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalse_ConstFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function test(string $s): void {
    echo $s;
}
test(false);
`)
	test.Expect = []string{
		"false passed to non-falseable parameter s in function test",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalse_ArrayDimFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function test(string $s): void {
    echo $s;
}
$arr = [false];
test($arr[0]);
`)
	test.Expect = []string{
		"false passed to non-falseable parameter s in function test",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalse_ListExpr(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function test(string $s): void {
    echo $s;
}
list($a) = [false];
test($a);
`)
	test.Expect = []string{
		"false passed to non-falseable parameter s in function test",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalse_PropertyFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public $prop = false;
}
function test(string $s): void {
    echo $s;
}
$a = new A();
test($a->prop);
`)
	test.Expect = []string{
		"false passed to non-falseable parameter s in function test",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalse_StaticCall(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class B {
    public static function getValue(): string|false {
        return false;
    }
}
function test(string $s): void {
    echo $s;
}
test(B::getValue());
`)
	test.Expect = []string{
		"false passed to non-falseable parameter s in function test",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalse_StaticPropertyFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class C {
    public static string|false $value = false;
}
function test(string $s): void {
    echo $s;
}
test(C::$value);
`)
	test.Expect = []string{
		"false passed to non-falseable parameter s in function test",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalse_FunctionCall(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function getValue(): string|false {
    return false;
}
function test(string $s): void {
    echo $s;
}
test(getValue());
`)
	test.Expect = []string{
		"false passed to non-falseable parameter s in function test",
	}
	test.RunAndMatch()
}

func TestNotEqualFalseCondition(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User {
    public function do(): bool {
    return false;
    }
}

/**
 * @return User|false
 */
function getUser():User|false {
    return false;
}

$b = getUser();

if ($b !== false){
$a = $b->do();
}
`)
	test.Expect = []string{
		"Missing PHPDoc for \\User::do public method",
	}
	test.RunAndMatch()
}

func TestNotEqualFalseElseCondition(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User {
    public function do(): bool {
    return false;
    }
}

/**
 * @return User|false
 */
function getUser():User|false {
    return false;
}

$b = getUser();

if ($b !== false){
$a = $b->do();
} else{
$a = $b->do();
}
`)
	test.Expect = []string{
		"Missing PHPDoc for \\User::do public method",
		"Call to undefined method",
		"potential false in b when accessing method",
		"Duplicated if/else actions",
	}
	test.RunAndMatch()
}

func TestAssignFalseMethodCall(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User {
    public function do(): bool {
    return false;
    }
}

/**
 * @return User|false
 */
function getUser():User|false {
    return false;
}

$a = getUser()->do();

`)
	test.Expect = []string{
		"Missing PHPDoc for \\User::do public method",
		"potential false in \\getUser when accessing method",
	}
	test.RunAndMatch()
}
