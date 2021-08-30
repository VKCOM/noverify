package regression_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestIssue362_1(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function method_exists($object, $method_name) { return 1 != 0; }

class Foo {
  public $value;
}

$x = new Foo();
if (method_exists($x, 'm1')) {
  $x->m1();
  $x->m1(1, 2);
}

if (method_exists(new Foo(), 'm2')) {
  if (method_exists(new Foo(), 'm3')) {
    (new Foo())->m2((new Foo())->m3());
  }
  (new Foo())->m2();
}

$y = new Foo();
if (method_exists($x, 'x1')) {
  $x->x1();
} elseif (method_exists($y, 'y1')) {
  $foo = $y->y1();
  if ($foo instanceof Foo) {
    echo $foo->value;
  }
}
`)
}

func TestIssue362_2(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
function method_exists($object, $method_name) { return 1 != 0; }

class Foo {}

$x = new Foo();
if (method_exists($x, 'm1')) {
}
$x->m1(); // Bad: called outside of if

if (method_exists(new Foo(), 'm2')) {
  if (method_exists(new Foo(), 'm3')) {
    $x->m2(); // Bad: called a method on a different object expression
  }
}
(new Foo())->m3(); // Bad: called outside of if

$y = new Foo();
if (method_exists($x, 'x1')) {
  $x->y1();
} elseif (method_exists($y, 'y1')) {
  $v = $y->x1();
  $v->foo();
}
`)
	test.Expect = []string{
		`Call to undefined method {\Foo}->m1()`,
		`Call to undefined method {\Foo}->m2()`,
		`Call to undefined method {\Foo}->m3()`,
		`Call to undefined method {\Foo}->y1()`,
		`Call to undefined method {\Foo}->x1()`,
		`Call to undefined method {mixed}->foo()`,
	}
	test.RunAndMatch()
}
