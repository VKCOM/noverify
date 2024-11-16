package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestShapePropertyFetch(t *testing.T) {
	// Make sure we don't allow to access shape elements via
	// property fetch syntax.

	test := linttest.NewSuite(t)
	test.AddNamedFile("a/b/test.php", `<?php
declare(strict_types = 1);
/**
 * @param \shape(x:int,y:float) $s
 */
function f($s) {
  $_ = $s->x;
  $_ = $s->y;
}
`)
	test.Expect = []string{
		`Property {shape{x:int,y:float}}->x does not exist`,
		`Property {shape{x:int,y:float}}->y does not exist`,
	}
	test.RunAndMatch()
}

func TestShapeDimFetch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
declare(strict_types = 1);
class Foo {
  public $x = 10;
  public $y = 1.5;
}

/** @return shape(foo:\Foo,err:string) */
function foo_shape() {
  return [];
}

function f() {
  $s = foo_shape();
  $_ = $s['foo']->x;
  $_ = $s['foo']->y;
  $_ = $s['bar']->z; // ['bar'] returns mixed type
}`)
	test.Expect = []string{`Property {mixed}->z does not exist`}
	test.RunAndMatch()
}

func TestShapeIntKey(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
declare(strict_types = 1);
class Box { public $value; }

function f() {
  /** @var array{0:int,1:\Box} $a1 */
  global $a1;

  $_ = $a1[0]->value; // Bad
  $_ = $a1[2]->value; // Bad
  $_ = $a1[1]->value;
  $_ = $a1['1']->value; // OK: int-like strings are casted to int keys by PHP
}
`)
	test.Expect = []string{
		`Property {int}->value does not exist`,
		`Property {mixed}->value does not exist`,
	}
	test.RunAndMatch()
}

func TestShapeSyntax(t *testing.T) {
	test := linttest.NewSuite(t)
	test.Config().StrictMixed = true
	test.AddFile(`<?php
declare(strict_types = 1);
class Box { public $value; }

class T {
  /** @var shape(x?:int) */
  public $good1;

  /** @var shape(*) */
  public $good2;

  /** @var shape(x:int, ...) */
  public $good3;

  /** @var shape{x:Box,y:Box} */
  public $good4;

  /** @var array{x:Box,y:Box} */
  public $good5;

  /** @var shape<x:Box,y:Box> */
  public $good6;

  /** @var shape(x[]:a) */
  public $bad1; // Bad: invalid shape element key.

  /** @var shape(x) */
  public $bad2; // Bad: invalid shape element.
}

function f() {
  /** @var shape(int) $a */
  global $a;
  $_ = $a;
}

$t = new T();
echo $t->bad1['x']->bad1;
echo $t->bad2['x']->bad2;

echo $t->good4['x']->value;
echo $t->good4['y']->value;
echo $t->good5['x']->value;
echo $t->good5['y']->value;
echo $t->good6['x']->value;
echo $t->good6['y']->value;
`)
	test.Expect = []string{
		`Invalid shape key: x[]`,
		`Shape param #1: want key:type, found x`,
		`Property {mixed}->bad1 does not exist`,
		`Property {mixed}->bad2 does not exist`,
		`Shape param #1: want key:type, found int`,
	}
	test.RunAndMatch()
}

func TestShapeReturn(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types = 1);
class MyList {
  /** @var \MyList */
  public $next;
}

/** @return shape(list:MyList) */
function new_shape1() { return []; }

$s1 = new_shape1();
echo $s1['list']->next->next;
`)
}

func TestTuple(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
declare(strict_types = 1);
class Box { public $value; }

class T {
  /** @var \tuple(int) */
  public $good1;

  /** @var tuple(Box, ...) */
  public $good2;

  /** @var tuple(int, \tuple(Box)) */
  public $good3;
}

$t = new T();
echo $t->good2[0]->value;
echo $t->good3[1][0]->value;
`)
}
