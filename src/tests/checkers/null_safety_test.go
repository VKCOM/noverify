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
