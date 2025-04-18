package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
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
		"potentially not safe call in function test signature of param",
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
		"potentially not safe access in parameter s of function test",
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
		"potentially not safe array access in parameter s of function test",
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
		"potentially not safe call in function test signature of param s",
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
		"potentially not safe accessing property 'value'",
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
		"potentially not safe call in function test signature of param s when calling function \\getValue",
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
		"potentially not safe call in b when accessing method",
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
		"potentially not safe call in \\getUser when accessing method",
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
		"potentially not safe access in parameter value of function testValue",
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
		"potentially not safe static call in function test signature of param s",
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
		"potentially not safe call in function testValue signature of param value when calling function \\falseFunc",
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
		"potentially not safe accessing property 'b'",
	}
	test.RunAndMatch()
}

func TestIsCondition(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
 class User {
    public string $name = "Kate";
}

function getUser(): User|int {
    return 1;
}

$x = getUser();
if (is_int($x)){
$a = $x->name;
} else{
$b = $x->name;
} 
`)
	test.Expect = []string{
		"Property {int}->name does not exist",
		"potentially not safe call when accessing property",
	}
	test.RunAndMatch()
}

func TestIsObjectCondition(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
 class User {
    public string $name = "Kate";
}

function getUser(): User|int {
    return 1;
}

$x = getUser();
if (is_object($x)){
$a = $x->name;
} else{
$b = $x->name;
}
`)
	test.Expect = []string{
		"Property {int}->name does not exist",
		"potentially not safe call when accessing property",
	}
	test.RunAndMatch()
}

func TestInheritDoc(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
   /**
     * Normalize path.
     *
     * @param string $path
     *
     *
     * @return string
     */
     function normalizePath($path)
    {

    }

interface FilesystemInterface
{
        /**
         * Check whether a file exists.
         *
         * @param string $path
         *
         * @return bool
         */
        public function has($path);
    }

class Filesystem implements FilesystemInterface
{
        /**
         * @inheritdoc
         */
        public function has($path)
        {
            $path = normalizePath($path);

            return strlen($path) === 0 ? false : (bool) $this->getAdapter()->has($path);
        }

        /**
         * Assert a file is present.
         *
         * @param string $path path to file
         *
         * @throws FileNotFoundException
         *
         * @return void
         */
        public function assertPresent($path)
        {
            if ($this->config->get('disable_asserts', false) === false && ! $this->has($path)) {
            }
        }
    }
`)
	test.Expect = []string{
		"Call to undefined method {\\Filesystem}->getAdapter()",
		"Property {\\Filesystem}->config does not exist",
	}
	test.RunAndMatch()
}

func TestForceInferring(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User{
    const IS_CALLABLE        = '!is_string(%s) && is_callable(%s)';
    const STRICT_IS_CALLABLE = 'is_object(%s) && is_callable(%s)';

    private function getCallable($variable = '$value')
    {
        $tpl = $this->strictCallables ? self::STRICT_IS_CALLABLE : self::IS_CALLABLE;

        return sprintf($tpl, $variable, $variable);
    }
}
`)
	test.Expect = []string{
		"Property {\\User}->strictCallables does not exist",
	}
	test.RunAndMatch()
}

func TestVariableCondition(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User {
    public string $name;

}

function getUser(): User|bool {
return null;
}

$user = getUser();

if ($user) {
    echo "User found: " . $user->name;
} else {
    echo "User not found." . $user->name;
}
`)
	test.Expect = []string{
		"potentially not safe call when accessing property",
		"Property {false}->name does not exist",
		"potentially not safe call when accessing property",
	}
	test.RunAndMatch()
}

func TestVariableNotCondition(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User {
    public string $name;

}

function getUser(): User|bool {
return null;
}

$user = getUser();

if (!$user) {
    echo "User found: " . $user->name;
} else {
    echo "User not found." . $user->name;
}
`)
	test.Expect = []string{
		"potentially not safe call when accessing property",
		"Property {false}->name does not exist",
		"potentially not safe call when accessing property",
	}
	test.RunAndMatch()
}

func TestSelfNewInstanceHandlerWithAbstract(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php

// Abstract base class for task handlers.
abstract class TaskHandler {
    // Common handler functionality could be placed here.
}

// Final class representing a Task.
final class Task {
    private string $type;
    private ?TaskHandler $handler;

    // Private constructor to enforce factory usage.
    private function __construct(string $type, ?TaskHandler $handler) {
        $this->type    = $type;
        $this->handler = $handler;
    }

    // Factory method for creating a task with a handler.
    public static function newWithHandler(string $type, TaskHandler $handler): self {
        return new self($type, $handler);
    }
}

// Concrete task handler extending the abstract TaskHandler.
class ConcreteTaskHandler extends TaskHandler {
    // Factory method for creating an instance of ConcreteTaskHandler.
    public static function create(): self {
        return new self();
    }
}

$handler = ConcreteTaskHandler::create();
$task    = Task::newWithHandler("example_task", $handler);

`)
	test.Expect = []string{
		"Missing PHPDoc for \\Task::newWithHandler public method",
		"Missing PHPDoc for \\ConcreteTaskHandler::create public method",
	}
	test.RunAndMatch()
}
