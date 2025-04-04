package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestFunctionPassingNull(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public string $value = 'Hello';
}

function test(A $a): void {
    echo $a->value;
}

test(null);
`)
	test.Expect = []string{
		"null passed to non-nullable parameter a in function test",
	}
	test.RunAndMatch()
}

func TestPropertyFetchNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public ?B $b = null;
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
		"potential null dereference when accessing property 'b'",
	}
	test.RunAndMatch()
}

func TestChainedPropertyAccessNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);

class A {
    public ?B $b = null;
}

class B {
    public C $c;
    public function __construct() {
        // Initialize property c with an instance of C.
        $this->c = new C();
    }
}

class C {
    public string $value = 'Hello';
}

/**
 * Function expecting an object of class C.
 */
function test(C $c): void {
    echo $c->value;
}

$a = new A();
// $a->b is null, so accessing $a->b->c leads to a potential null dereference.
test($a->b->c);
`)
	test.Expect = []string{
		"potential null dereference when accessing property",
	}
	test.RunAndMatch()
}

// After fix issue with not correct checking existing variable after unpacking - should fail
func TestListExprNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public string $value = 'Hello';
}

function test(A $a): void {
    echo $a->value;
}

list($m, list($n, $o)) = [new A(), [new A(), null]];

test($m);
test($n); 
test($o);
`)
	test.Expect = []string{
		"not null safety call in function test signature of param",
		"Cannot find referenced variable $o",
		"Cannot find referenced variable $n",
	}
	test.RunAndMatch()
}

func TestArrayDimFetchNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public string $value = 'Hello';
}

function test(A $a): void {
    echo $a->value;
}

$arr = [new A(), null];
test($arr[1]);
`)
	test.Expect = []string{
		"potential null array access in parameter a of function test",
	}
	test.RunAndMatch()
}

func TestVariadicParameterNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public string $value = 'Hello';
}

/**
 * Variadic function accepting only objects of class A.
 */
function testVariadic(A ...$a): void {
    foreach ($a as $item) {
        echo $item->value, "\n";
    }
}

testVariadic(new A(), null);
`)
	test.Expect = []string{
		"null passed to non-nullable parameter a in function testVariadic",
	}
	test.RunAndMatch()
}

func TestIfNullCheckSafe(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);

class A {
    public string $value = 'Safe';
}

/**
 * Function expecting a non-null instance of A.
 */
function test(A $a): void {
    echo $a->value;
}

$v = null;
if ($v == null) {
    // Correctly assign a new instance when $v is null.
    $v = new A();
	test($v);
} else {
test($v);
}
test($v);
`)

	test.Expect = []string{
		`not null safety call in function test signature of param`,
		`not null safety call in function test signature of param`,
	}
	test.RunAndMatch()
}

func TestIfNullCheckUnsafe(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);

class A {
    public string $value = 'Unsafe';
}

/**
 * Function expecting a non-null instance of A.
 */
function test(A $a): void {
    echo $a->value;
}

$v = null;
if ($v == null) {
    // No assignment is done here – $v remains null.
}
test($v); // Should trigger a null safety error.
`)
	test.Expect = []string{
		"not null safety call in function test signature of param",
	}
	test.RunAndMatch()
}

func TestIfNotNullCheck(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
declare(strict_types=1);

class A {
    public string $value = 'NotNull';
}

/**
 * Function expecting a non-null instance of A.
 */
function test(A $a): void {
    echo $a->value;
}

$v = new A();
if ($v !== null) {
    // $v is known to be non-null.
    test($v); // Should be safe.
}
`)

	test.Expect = []string{}
	test.RunAndMatch()
}

func TestStaticNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public static $value = 'test';
    
    public static function hello(): void {
        echo "Hello!";
    }
}

$maybeClass = rand(0, 1) ? 'A' : null;
$maybeClass::hello();
echo $maybeClass::$value;
`)
	test.Expect = []string{
		"Missing PHPDoc for \\A::hello public method",
		"potential null dereference when accessing static call throw $maybeClass",
		"attempt to access property that can be null",
	}
	test.RunAndMatch()
}

func TestFetchPropertyNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User {
    public $name = "lol";
}

$user = new User();
$user = null;
echo $user->name;
`)
	test.Expect = []string{
		"potential attempt to access property through null",
	}
	test.RunAndMatch()
}

func TestStaticPropertyNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public static $value = null;
    
    public static function hello(): void {
        echo "Hello!";
    }
}

function test(string $a): void {
echo "test";
}
$maybeClass  = new A();
test($maybeClass::$value);

`)
	test.Expect = []string{
		"Missing PHPDoc for \\A::hello public method",
		`potential null dereference when accessing property 'value'`,
	}
	test.RunAndMatch()
}

func TestStaticPropertyFetchWithName(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public static string $value = 'test';
}

function test(string $s): void {
    echo $s;
}

test(A::$value);
`)
	test.Expect = []string{}
	test.RunAndMatch()
}

func TestStaticPropertyNullFetchWithName(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public static string $value = null;
}

function test(string $s): void {
    echo $s;
}

test(A::$value);
`)
	test.Expect = []string{
		"potential null dereference when accessing property 'value'",
	}
	test.RunAndMatch()
}

func TestStaticCallNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public static function hello(): ?string {
        return "Hello!";
    }
}

function test(string $s): void {
    echo $s;
}

test(A::hello());
`)
	test.Expect = []string{
		"Missing PHPDoc for \\A::hello public method",
		"not null safety call in function test signature of param s when calling static function hello",
	}
	test.RunAndMatch()
}

func TestFuncCallNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function nullFunc(){
return null;

function testValue(string $value): void {
    echo $value;
}

testValue(nullFunc());


}
`)
	test.Expect = []string{
		"not null safety call in function testValue signature of param value when calling function \\nullFunc",
		"Unreachable code",
	}
	test.RunAndMatch()
}

func TestStaticCallNullSafetyThroughVariable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public static function hello(): ?string {
        return "Hello!";
    }
}

function test(string $s): void {
    echo $s;
}

$maybeClass = new A();
test($maybeClass::hello());
`)
	test.Expect = []string{
		"Missing PHPDoc for \\A::hello public method",
		"not null safety call in function test signature of param s when calling static function hello",
	}
	test.RunAndMatch()
}

func TestFunctionCallNullSafetyThroughVariable(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
    public static function hello(): ?string {
        return "Hello!";
    }
}

function test(A $s): void {
    echo $s;
}

function testNullable(): ?A{
	return new A();
}

test(testNullable());
`)
	test.Expect = []string{
		"Missing PHPDoc for \\A::hello public method",
		"not null safety call in function test signature of param s when calling function \\testNullable",
	}
	test.RunAndMatch()
}

func TestVariableNotConditionNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User {
    public string $name;

}

function getUser(): User|null {
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
		"potential attempt to access property through null",
	}
	test.RunAndMatch()
}

func TestVariableInConditionNullSafety(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class User {
    public string $name;

}

function getUser(): User|null {
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
		"potential attempt to access property through null",
	}
	test.RunAndMatch()
}
