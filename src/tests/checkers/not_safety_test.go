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
		"not safety call in function test signature of param",
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
		"potential not safety access in parameter s of function test",
	}
	test.RunAndMatch()
}

func TestFunctionPassingArrayDimFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
/**
 * @param bool[] $boolArray Массив булевых значений
 */
function test(array $boolArray): void {
    echo $boolArray;
}
$arr = [false];
test($arr);
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestFunctionPassingArrayElemDimFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

function test(string $s): void {
    echo $s;
}
$arr = [false];
test($arr[0]);
`)
	test.Expect = []string{
		"not safety array access in parameter s of function test",
	}
	test.RunAndMatch()
}

func TestFunctionParamBoolArrayElemDimFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

function test(bool $s): void {
    echo $s;
}
$arr = [false];
test($arr[0]);
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestFunctionPassingListExprUnpack(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function test(string $s): void {
    echo $s;
}
list($a) = [false];
test($a);
`)
	test.Expect = []string{
		"not safety call in function test signature of param s",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalseStaticPropertyFetch(t *testing.T) {
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
		"potential not safety accessing property 'value'",
	}
	test.RunAndMatch()
}

func TestFunctionPassingFalseFunctionCall(t *testing.T) {
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
		"not safety call in function test signature of param s when calling function \\getValue",
	}
	test.RunAndMatch()
}

func TestNotIdenticalFalseCondition(t *testing.T) {
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
$с = $b->do();
}
`)
	test.Expect = []string{
		"Missing PHPDoc for \\User::do public method",
		"Call to undefined method",
		"potential not safety call in b when accessing method",
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
		"potential not safety call in \\getUser when accessing method",
	}
	test.RunAndMatch()
}

func TestFalseParamInFunc(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function testValue(string $value): void {
    echo $value;
}
testValue(false);
`)
	test.Expect = []string{
		"potential not safety access in parameter value of function testValue",
	}
	test.RunAndMatch()
}

func TestStaticCallNotSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public static function hello(): bool {
        return false;
    }
}

function test(string $s): void {
    echo $s;
}

test(A::hello());
`)
	test.Expect = []string{
		"Missing PHPDoc for \\A::hello public method",
		"not safety static call in function test signature of param s",
	}
	test.RunAndMatch()
}

func TestFuncCallSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function falseFunc(){
return false;

function testValue(string $value): void {
    echo $value;
}

testValue(falseFunc());


}
`)
	test.Expect = []string{
		"not safety call in function testValue signature of param value when calling function \\falseFunc",
		"Unreachable code",
	}
	test.RunAndMatch()
}

func TestPropertyFetchNotSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public B|false $b = false;
}
class B {
    public string $value = 'Test';
}

function test(B $b): void {
    echo $b->value;
}

$a = new A();
test($a->b);
`)
	test.Expect = []string{
		"potential not safety accessing property 'b'",
	}
	test.RunAndMatch()
}
