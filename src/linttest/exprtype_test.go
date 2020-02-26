package linttest_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/solver"
)

// Tests in this file make it less likely that type solving will break
// without being noticed.

// TODO(quasilyte): better handling of an `empty_array` type.
// Now it's resolved to `mixed[]` for expressions that have multiple empty_array.

func TestExprTypeMagicCall(t *testing.T) {
	tests := []exprTypeTest{
		{`$m->magic()`, `\Magic`},
		{`$m->magic()->f2()`, `\Magic`},
		{`$m->f2()->magic()`, `\Magic`},
		{`(new Magic())->magic()`, `\Magic`},
		{`$m->notMagic()`, `int`},
		{`$m->magic()->notMagic()`, `int`},
		{`$m->m1()->m2()->notMagic()`, `int`},

		{`$m2->unknown()`, `mixed`},
		{`$m2->magicInt()`, `int`},
		{`$m2->magicString()`, `string`},
		{`$m2->add(1, 2)`, `int`},
		{`Magic2::getInstance()->magicInt()`, `int`},
		{`Magic2::unknown()`, `mixed`},

		// @method annotations should take precedence over
		// generic __call return type info.
		{`$m3->magicInt()`, `int`},
		{`$m3->unknown()`, `\Magic3`},
		{`$m3->magic()->magicInt()`, `int`},

		{`StaticMagic::magicInt()`, `int`},
		{`StaticMagic::newMagic()`, `\Magic`},
		{`StaticMagic::magic()->magic()`, `\Magic`},
	}

	global := `<?php
class Magic {
  public function __call() { return $this; }
  public function notMagic() { return 10; }
}

/**
 * @method int magicInt()
 * @method string magicString()
 * @method int add(int $x, int $y)
 * @method static Magic2 getInstance()
 */
class Magic2 {}

/**
 * @method int magicInt
 */
class Magic3 {
  public function __call() { return $this; }
}

/**
 * @method static int magicInt()
 */
class StaticMagic {
  public function __callStatic() { return new Magic(); }
}
`
	local := `$m = new Magic(); $m2 = new Magic2(); $m3 = new Magic3();`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeRef(t *testing.T) {
	tests := []exprTypeTest{
		{`$v =& $ints[0]`, `mixed`},
		{`assign_ref_dim_fetch1()`, `mixed[]`},
		{`assign_ref_dim_fetch2()`, `mixed[]`},
		{`assign_ref_dim_fetch3()`, `mixed[]`},
	}
	global := `<?php
$ints = [1, 2];

function assign_ref_dim_fetch1() {
  global $ints;
  $x[] =& $ints;
  return $x;
}

function assign_ref_dim_fetch2() {
  global $ints;
  $x[] =& $ints[0];
  return $x;
}

function assign_ref_dim_fetch3() {
  global $ints;
  $x[0][] =& $ints[0];
  return $x;
}
`
	runExprTypeTest(t, &exprTypeTestContext{global: global}, tests)
}

func TestExprTypeFixes(t *testing.T) {
	tests := []exprTypeTest{
		{`alias_double()`, `float`},
		{`alias_real()`, `float`},
		{`alias_integer()`, `int`},
		{`alias_long()`, `int`},
		{`alias_boolean()`, `bool`},
		{`untyped_array()`, `mixed[]`},
		{`dash()`, `mixed`},
		{`array1()`, `int[]`},
		{`array2()`, `int[][]`},
		{`array_int()`, `int[]`},
		{`array_int_string()`, `string[]`},      // key type is currently ignored
		{`array_int_stdclass()`, `\stdclass[]`}, // key type is currently ignored
		{`array_return_string()`, `string`},
	}

	global := `<?php
/** @return real */
function alias_real() {}

/** @return double */
function alias_double() {}

/** @return integer */
function alias_integer() {}

/** @return long */
function alias_long() {}

/** @return boolean */
function alias_boolean() {}

/** @return [] */
function untyped_array() {}

/** @return - some result */
function dash() {}

/** @return []int */
function array1() {}

/** @return [][]int */
function array2() {}

/** @return array<int> */
function array_int() {}

/** @return array<int, string> */
function array_int_string() {}

/** @return array<int, stdclass> */
function array_int_stdclass() {}

/** @param array<int,string> $a */
function array_return_string($a) { return $a[0]; }
`

	runExprTypeTest(t, &exprTypeTestContext{global: global}, tests)
}

func TestExprTypeVoid(t *testing.T) {
	tests := []exprTypeTest{
		{`void_func1()`, `void`},
		{`void_func2()`, `void`},
		{`void_func3()`, `void`},
		{`$foo->voidMeth1()`, `void`},
		{`$foo->voidMeth2()`, `void`},
		{`$foo->voidMeth3()`, `void`},
	}

	global := `<?php
function void_func1() {
  echo 123;
}

function void_func2() { return; }

/** @return void */
function void_func3() {}

class Foo {
  public function voidMeth1() {}
  public function voidMeth2() { return; }

  /** @return void */
  public function voidMeth3() {}
}
`
	local := `$foo = new Foo();`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeArrayAccess(t *testing.T) {
	tests := []exprTypeTest{
		{`$ints[0]`, `int`},
		{`getInts()[0]`, `int`},
		{`$self[0]`, `\Self`},
		{`$self[0][1]`, `\Self`},
		{`$self[0][1]->offsetGet(2)`, `\Self`},
	}

	global := `<?php
function getInts() { return new Ints(); }

class Ints implements ArrayAccess {
   /** @return bool */
   public function offsetExists($offset) {}
   /** @return int */
   public function offsetGet($offset) {}
   /** @return void */
   public function offsetSet($offset, $value) {}
   /** @return void */
   public function offsetUnset($offset) {}
}

class Self implements ArrayAccess {
   /** @return bool */
   public function offsetExists($offset) {}
   /** @return Self */
   public function offsetGet($offset) {}
   /** @return void */
   public function offsetSet($offset, $value) {}
   /** @return void */
   public function offsetUnset($offset) {}
}
`
	local := `$ints = new Ints(); $self = new Self();`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeAnnotatedProperty(t *testing.T) {
	tests := []exprTypeTest{
		{`$x->int`, `int`},
		{`$x->getInt()`, `int`},
	}

	global := `<?php
/**
 * @property int $int optional description
 */
class Foo {
  /***/
  public function getInt() {
    return $this->int;
  }
}`
	local := `$x = new Foo();`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeMalformedPhpdoc(t *testing.T) {
	tests := []exprTypeTest{
		{`return_mixed(0)`, `mixed`},
		{`return_int(0)`, `int`},
		{`return_int2(0)`, `int`},
		{`return_int3(0)`, `int`},
	}

	global := `<?php
/**
 * @param int &$x
 */
function return_int2(&$x) { return $x; }

/**
 * @param int &$x
 */
function return_int3($x) { return $x; }

/**
 * @param $x
 */
function return_mixed($x) { return $x; }

/**
 * @param int
 */
function return_int($x) { return $x; }
`
	runExprTypeTest(t, &exprTypeTestContext{global: global}, tests)
}

func TestExprTypeMagicGet(t *testing.T) {
	tests := []exprTypeTest{
		{`(new Ints)->a`, `int`},
		{`$ints->a`, `int`},
		{`$ints->b`, `int`},
		{`(new Chain)->chain`, `\Chain`},
		{`$chain->chain`, `\Chain`},
		{`$chain->chain->chain`, `\Chain`},
	}

	global := `<?php
class Ints {
  public function __get($k) { return 0; }
}
class Chain {
  public function __get($k) { return $this; }
}`
	local := `
$ints = new Ints();
$chain = new Chain();`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeLateStaticBinding(t *testing.T) {
	tests := []exprTypeTest{
		{`getBase()`, `\Base`},
		{`getDerived()`, `\Base|\Derived`},
		{`getBase2()`, `\Base`},
		{`getDerived2()`, `\Base|\Derived`},
		{`getBase2()->getStatic()->getStatic()`, `\Base`},
		{`getDerived2()->getStatic()->getStatic()`, `\Base|\Derived`},
		{`eitherDerived()`, `\Derived|\DerivedDerived`},
		{`eitherDerived()->getStatic()`, `\Base|\Derived|\DerivedDerived`},

		{`Base::staticNewStatic()`, `\Base`},
		{`Base::staticNewStatic()->staticNewStatic()`, `\Base`},
		{`Derived::staticNewStatic()`, `\Derived`},
		{`Derived::staticNewStatic()->staticNewStatic()`, `\Derived`},
		{`DerivedDerived::staticNewStatic()`, `\DerivedDerived`},
		{`DerivedDerived::staticNewStatic()->staticNewStatic()`, `\DerivedDerived`},

		{`$b->newStatic()`, `\Base`},
		{`$d->newStatic()`, `\Derived`},
		{`$dd->newStatic()`, `\DerivedDerived`},

		{`$b->getStatic()`, `\Base`},
		{`$b->getStatic()->getStatic()`, `\Base`},
		{`$b->getStaticArray()`, `\Base[]`},
		{`$b->getStaticArray()[0]`, `\Base`},
		{`$b->getStaticArrayArray()`, `\Base[][]`},
		{`$b->getStaticArrayArray()[0][0]`, `\Base`},

		{`$d->getStatic()`, `\Base|\Derived`},
		{`$d->getStatic()->getStatic()`, `\Base|\Derived`},
		{`$d->getStaticArray()`, `\Derived[]`},
		{`$d->getStaticArray()[0]`, `\Derived`},
		{`$d->getStaticArrayArray()`, `\Derived[][]`},
		{`$d->getStaticArrayArray()[0][0]`, `\Derived`},

		{`$dd->getStatic()`, `\Base|\DerivedDerived`},
		{`$dd->getStatic()->getStatic()`, `\Base|\DerivedDerived`},
		{`$dd->getStaticArray()`, `\DerivedDerived[]`},
		{`$dd->getStaticArray()[0]`, `\DerivedDerived`},
		{`$dd->getStaticArrayArray()`, `\DerivedDerived[][]`},
		{`$dd->getStaticArrayArray()[0][0]`, `\DerivedDerived`},

		{`$b->initAndReturnOther1()`, `\Base`},
		{`$b->initAndReturnOther2()`, `\Base`},

		{`(new Base())->getStatic()`, `\Base`},
		{`(new Derived())->getStatic()`, `\Base|\Derived`},

		{`$d->derivedGetStatic()`, `\Derived`},
		{`$d->derivedNewStatic()`, `\Derived`},
		{`$dd->derivedGetStatic()`, `\Derived|\DerivedDerived`},
		{`$dd->derivedNewStatic()`, `\DerivedDerived`},

		{`$d->getStatic()`, `\Base|\Derived`},
		{`$d->getStatic()->getStatic()`, `\Base|\Derived`},
		{`$dd->getStatic()`, `\Base|\DerivedDerived`},
		{`$dd->getStatic()->getStatic()`, `\Base|\DerivedDerived`},

		{`$d->getStaticForOverride1()`, `null|\Derived`},
		{`$d->getStaticForOverride2()`, `\Derived`},
		{`$d->getStaticForOverride3()`, `\Derived`},
		{`$dd->getStaticForOverride1()`, `null|\DerivedDerived`},
		{`$dd->getStaticForOverride2()`, `\Derived`}, // Since $this works like `self`
		{`$dd->getStaticForOverride3()`, `\Derived|\DerivedDerived`},

		{`$dd->asParent()`, `\Derived|\DerivedDerived`},
		{`$dd->asParent()->newStatic()`, `\Derived|\DerivedDerived`},
		{`$dd->asParent()->asParent()`, `\Derived|\DerivedDerived`},

		// Resolving of `$this` (which should be identical to `static`).
		{`$b->getThis()`, `\Base`},
		{`$d->getThis()`, `\Base|\Derived`},
		{`$b->getThis()->getThis()`, `\Base`},
		{`$d->getThis()->getThis()`, `\Base|\Derived`},

		// TODO: resolve $this without @return hint into `static` as well?
		{`$b->getThisNoHint()`, `\Base`},
		{`$d->getThisNoHint()`, `\Base`},
		{`$dd->getThisNoHint()`, `\Base`},
	}

	global := `<?php
class Base {
  /** @return $this */
  public function getThis() { return $this; }

  public function getThisNoHint() { return $this; }

  /** @return static */
  public function getStatic() { return $this; }

  /** @return static[] */
  public function getStaticArray($x) { return []; }

  /** @return static[][] */
  public function getStaticArrayArray($x) { return []; }

  /** Doesn't require return type hint */
  public function newStatic() { return new static(); }

  /** @return static */
  public function getStaticForOverride1() { return $this; }

  /** @return static */
  public function getStaticForOverride2() { return $this; }

  /** @return static */
  public function getStaticForOverride3() { return $this; }

  public static function staticNewStatic() { return new static(); }

  public function initAndReturnOther1() {
    $this->other1 = new static();
    return $this->other1;
  }

  public function initAndReturnOther2() {
    $other2 = new static();
    return $other2;
  }

  /** @var static */
  public $other1;
}

class Derived extends Base {
  /** @return static */
  public function derivedNewStatic() { return new static(); }

  /** @return static */
  public function derivedGetStatic() { return $this; }

  /** @return static */
  public function getStaticForOverride1() { return null; }

  public function getStaticForOverride2() { return $this; }

  /** @return $this */
  public function getStaticForOverride3() { return $this; }
}

class DerivedDerived extends Derived {
  /** @return Derived */
  public function asParent() { return $this; }
}

function getBase() {
  return (new Base())->getStatic();
}

function getDerived() {
  return (new Derived())->getStatic();
}

function getBase2() {
  $b = new Base();
  $b2 = $b->getStatic();
  return $b2;
}

function getDerived2() {
  $d = new Derived();
  $d2 = $d->getStatic();
  return $d2;
}

function eitherDerived($cond) {
  if ($cond) {
    return new Derived();
  }
  return new DerivedDerived();
}
`

	local := `
$b = new Base();
$d = new Derived();
$dd = new DerivedDerived();
`

	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeSimple(t *testing.T) {
	tests := []exprTypeTest{
		{`true`, "bool"},
		{`false`, "bool"},
		{`(bool)1`, "bool"},
		{`(boolean)1`, "bool"},

		{`1`, "int"},
		{`(int)1.5`, "int"},
		{`(integer)1.5`, "int"},

		{`1.21`, "float"},
		{`(float)1`, "float"},
		{`(real)1`, "float"},
		{`(double)1`, "float"},

		{`""`, "string"},
		{`(string)1`, "string"},

		{`[]`, "mixed[]"},
		{`[1, "a", 4.5]`, "mixed[]"},

		{`1+5<<2`, `int`},

		{`-1`, `int`},
		{`-1.4`, `float`},
		{`+1`, `int`},
		{`+1.4`, `float`},

		{`$int`, "int"},
		{`$float`, "float"},
		{`$string`, "string"},

		{`define('foo', 0 == 0)`, `void`},
	}

	global := `<?php
function define($name, $value) {}
define('true', (bool)1);
define('false', (bool)0);
$int = 10;
$float = 20.5;
$string = "123";
`
	runExprTypeTest(t, &exprTypeTestContext{global: global}, tests)
}

func TestExprTypeArray(t *testing.T) {
	tests := []exprTypeTest{
		{`[]`, `mixed[]`}, // Should never be "empty_array" after resolving
		{`[[]]`, `mixed[]`},

		{`[1, 2]`, "int[]"},
		{`[1.4, 3.5]`, "float[]"},
		{`["1", "5"]`, "string[]"},

		{`["k1" => 123, "k2" => 345]`, `int[]`},
		{`[0 => "a", 1 => "b"]`, `string[]`},

		{`[$int, $int]`, "mixed[]"}, // TODO: could be int[]

		{`$ints[0]`, "int"},
		{`["11"][0]`, "string"},
		{`[1.4][0]`, "float"},
	}

	local := `$int = 10; $ints = [1, 2];`
	runExprTypeTest(t, &exprTypeTestContext{local: local}, tests)
}

func TestExprTypeMulti(t *testing.T) {
	tests := []exprTypeTest{
		{`$cond ? 1 : 2`, "int"},
		{`$int_or_float`, "int|float"},
		{`$int_or_float`, "float|int"},
		{`$cond ? 10 : "123"`, "int|string"},
		{`$cond ? ($int_or_float ? 10 : 10.4) : (bool)1`, "int|float|bool"},
		{`$bool_or_int`, `bool|int`},
		{`$cond ? 10 : get_mixed(1)`, `int|mixed`},
		{`$cond ? get_mixed(1) : 10`, `int|mixed`},
	}

	global := `<?php
/** @return mixed */
function get_mixed($x) { return $x; }

$cond = "true";
$int_or_float = 10;
if ($cond) {
  $int_or_float = 10.5;
}
`
	local := `
/** @var bool|int $bool_or_int */
$bool_or_int = 10;`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeOps(t *testing.T) {
	tests := []exprTypeTest{
		{`1 + $int`, "int"},
		{`$int + 1`, "int"},
		{`1 + (int)$float`, "int"},
		{`1 + $global_int`, "float"},
		{`$global_int + 1`, "float"},
		{`1 + $float`, "float"},

		{`$int . $float`, "string"},

		{`$int && $float`, "bool"},
		{`$int || 1`, "bool"},
	}

	global := `<?php
$global_int = 10;
$global_float = 20.5;`
	local := `
$int = 10;
$float = 20.5;
$string = "123";
$bool = (bool)1;`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeProperty(t *testing.T) {
	tests := []exprTypeTest{
		{`$point->x`, "float"},
		{`$point->y`, "float"},

		{`Gopher::$name`, "string"},
		{`Gopher::POWER`, "int"},
		{`$magic->int`, "int"},
	}

	global := `<?php

class Gopher {
  /** @var string */
  public static $name = "unnamed";

  const POWER = 9001; // It's over 9000
}

/**
 * @property int $int
 */
class Magic {
  public function __get($prop_name) {}
}

class Point {
  /** @var float */
  public $x;
  /** @var float */
  public $y;
}
`
	local := `
$point = new Point();
$magic = new Magic();`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeFunction(t *testing.T) {
	tests := []exprTypeTest{
		{`get_ints()`, `int[]`},
		{`get_floats()`, `float[]`},
		{`get_array()`, `mixed[]`},
		{`get_array_or_null()`, `mixed[]|null`},
		{`get_null_or_array()`, `mixed[]|null`},
		{`try_catch1()`, `bool|int|string`},
		{`try_finally1()`, `bool|int|string`},
		{`ifelse1()`, `bool|int|string`},
		{`ifelse2()`, `bool|int|string`},
		{`ifelse3()`, `bool|int|string`},
		{`switch1()`, `bool|int|string`},
		{`switch2()`, `bool|int|string`},
		{`switch3()`, `bool|string`},
		{`throw1()`, `int`},
		{`throw2()`, `bool|int`},
		{`foreach1()`, `int|string`},
		{`foreach2()`, `int|string`},
		{`undefined_type1()`, `mixed`},
		{`undefined_type2()`, `mixed`},
		{`untyped_param()`, `mixed`},
		{`bare_ret1()`, `int|null`},
		{`bare_ret2()`, `int|null`},
		{`recur1()`, `int|string`},
		{`recur2()`, `int|string`},
		{`recur3()`, `mixed`},
		{`recur4()`, `mixed`},
		{`recur5()`, `mixed`},
		{`mixed_array()`, `mixed[]`},
		{`mixed_or_ints1()`, `mixed[]|int[]`},
		{`mixed_or_ints2()`, `mixed[]|int[]`},
		{`mixed_array()[1]`, `mixed`},
		{`mixed_or_ints1()[1]`, `mixed|int`},
		{`mixed_or_ints2()[1]`, `mixed|int`},
	}

	global := `<?php
function define($name, $value) {}
define('null', 0);

class Foo {}

function mixed_array($x) {
  return [$x, 1, 2];
}

function mixed_or_ints1($x) {
  if ($x) {
    return mixed_array($x);
  }
  return [0, 0];
}

function mixed_or_ints2($x) {
  $a = array(0, 0);
  if ($x) {
    $a = mixed_array($x);
  }
  return $a;
}

function recur1($cond) {
  if ($cond) { return 0; }
  return recur2($cond);
}

function recur2($cond) {
  if ($cond) { return ""; }
  return recur1($cond);
}

function recur3() { return recur4(); }
function recur4() { return recur5(); }
function recur5() { return recur3(); }

function bare_ret1($cond) {
  if ($cond) { return; }
  return 10;
}

function bare_ret2($cond) {
  if ($cond) { return 10; }
  return;
}

function untyped_param($x) { return $x; }

function undefined_type1() {
  $x = unknown_func();
  return $x;
}

function undefined_type2() {
  return $x;
}

function foreach1($xs) {
  foreach ($xs as $_) {
    return 10;
  }
  return "";
}

function foreach2($xs, $cond) {
  foreach ($xs as $_) {
    if ($cond[0]) {
      if ($cond[1]) {
        return 10;
      }
    }
  }
  return "";
}

function throw1($cond) {
  if ($cond) {
    return 10;
  }
  throw new Exception();
}

function throw2($cond) {
  if ($cond[0]) {
    throw new Exception("");
  } else if ($cond[1]) {
    return 10;
  } else if ($cond[2]) {
    throw new Exception("");
  } else if ($cond[3]) {
    return false;
  }
  throw new Exception("");
}

function get_ints() {
	$a = []; // "empty_array"
	$a[0] = 1;
	$a[1] = 2;
	return $a; // Should be resolved to just int[]
}

function switch1($v) {
  switch ($v) {
  case 10:
    return 10;
  case 20:
    return "";
  default:
    return false;
  }
}

function switch2($v) {
  switch ($v) {
  case 10:
    return 10;
  case 20:
    return "";
  }
  return false;
}

function switch3($v) {
  switch ($v) {
  default:
    return "";
  }
  return false;
}

function ifelse1($cond) {
  if ($cond) {
    return 10;
  } else if ($cond+1) {
    return "";
  } else {
    return false;
  }
}

function ifelse2($cond) {
  if ($cond) {
    return 10;
  } elseif ($cond+1) {
    return "";
  } else {
    return false;
  }
}

function ifelse3($cond) {
  if ($cond) {
    return 10;
  } elseif ($cond+1) {
    return "";
  }
  return false;
}

function try_catch1() {
  try {
    return 10;
  } catch (Exception $_) {
    return "";
  }
  return false;
}

function try_finally1() {
  try {
    return 10;
  } finally {
    return "";
  }
  return false;
}

/** @return float[] */
function get_floats() { return []; }

function get_array() { return []; }

/** @return array */
function get_array_or_null() { return null; }

/** @return null */
function get_null_or_array() { return []; }`
	runExprTypeTest(t, &exprTypeTestContext{global: global}, tests)
}

func TestExprTypeMethod(t *testing.T) {
	tests := []exprTypeTest{
		{`\NS\Test::instance()`, `\NS\Test`},
		{`\NS\Test::instance2()`, `\NS\Test`},
		{`$test->getInt()`, `int`},
		{`$test->getInts()`, `int[]`},
		{`$test->getThis()->getThis()->getInt()`, `int`},
		{`new \NS\Test()`, `\NS\Test`},
	}

	global := `<?php
namespace NS {
	class Test {
		public function getInt() { return 10; }
		public function getInts() { return [1, 2]; }
		public function getThis() { return $this; }

		public static function instance() {
			return self::$instances[0];
		}

		public static function instance2() {
			foreach (self::$instances as $instance) {
				return $instance;
			}
		}

		/** @var Test[] */
		public static $instances;
	}
}`
	local := `$test = new \NS\Test(); $derived = new Derived();`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeInterface(t *testing.T) {
	tests := []exprTypeTest{
		{"$foo", `\Foo`},
		{"$foo->getThis()", `\Foo`},
		{"$foo->acceptThis($foo)", `\TestInterface`},
		{"$foo->acceptThis($foo)->acceptThis($foo)", `\TestInterface`},
	}

	global := `<?php
interface TestInterface {
  /**
   * @return self
   */
  public function getThis();

  /**
   * @param \TestInterface $x
   * @return \TestInterface
   */
  public function acceptThis($x);
}

class Foo implements TestInterface {
  public function getThis() { return $this; }

  public function acceptThis($x) { return $x->getThis(); }
}`
	local := `$foo = new Foo();`
	runExprTypeTest(t, &exprTypeTestContext{global: global, local: local}, tests)
}

func TestExprTypeOverride(t *testing.T) {
	tests := []exprTypeTest{
		{`array_shift($ints)`, "int"},
		{`array_slice($ints, 0, 2)`, "int[]"},
		{`array_map(function($x) { return $x; }, $ints)`, `mixed[]`},
	}

	stubs := `<?php
/**
 * @param callable $callback
 * @param array $arr1
 * @param array $_ [optional]
 * @return array an array containing all the elements of arr1
 */
function array_map($callback, array $arr1, array $_ = null) { }

/**
 * @param array $array
 * @param int $offset
 * @param int $length [optional]
 * @param bool $preserve_keys [optional]
 * @return array the slice.
 */
function array_slice (array $array, $offset, $length = null, $preserve_keys = false) {}

/**
 * @param array $array
 * @return mixed the shifted value, or &null; if array is
 * empty or is not an array.
 */
function array_shift (array &$array) {}

namespace PHPSTORM_META {
	override(\array_slice(0), type(0));
	override(\array_shift(0), elementType(0));
}`
	local := `$ints = [1, 2];`
	runExprTypeTest(t, &exprTypeTestContext{stubs: stubs, local: local}, tests)
}

func runExprTypeTest(t *testing.T, ctx *exprTypeTestContext, tests []exprTypeTest) {
	if ctx == nil {
		ctx = &exprTypeTestContext{}
	}

	meta.ResetInfo()
	if ctx.stubs != "" {
		linttest.ParseTestFile(t, "stubs.php", ctx.stubs)
		meta.Info.InitStubs()
	}
	var gw globalsWalker
	if ctx.global != "" {
		if !strings.HasPrefix(ctx.global, "<?php") {
			t.Error("missing <?php tag in global PHP code snippet")
			return
		}
		root, _ := linttest.ParseTestFile(t, "exprtype_global.php", ctx.global)
		root.Walk(&gw)
	}
	sources := exprTypeSources(ctx, tests, gw.globals)
	linttest.ParseTestFile(t, "exprtype.php", sources)
	meta.SetIndexingComplete(true)
	linttest.ParseTestFile(t, "exprtype.php", sources)

	for i, test := range tests {
		fn, ok := meta.Info.GetFunction(fmt.Sprintf("\\f%d", i))
		if !ok {
			t.Errorf("missing f%d info", i)
			continue
		}
		have := solver.ResolveTypes("", fn.Typ, make(map[string]struct{}))
		want := makeType(test.expectedType)
		if !reflect.DeepEqual(have, want) {
			t.Errorf("type mismatch for %q:\nhave: %q\nwant: %q",
				test.expr, have, want)
		}
	}
}

func makeType(typ string) map[string]struct{} {
	if typ == "" {
		return map[string]struct{}{}
	}

	res := make(map[string]struct{})
	for _, t := range strings.Split(typ, "|") {
		res[t] = struct{}{}
	}
	return res
}

type exprTypeTest struct {
	expr         string
	expectedType string
}

type exprTypeTestContext struct {
	global string
	local  string
	stubs  string
}

func exprTypeSources(ctx *exprTypeTestContext, tests []exprTypeTest, globals []string) string {
	var buf strings.Builder
	buf.WriteString("<?php\n")
	for i, test := range tests {
		fmt.Fprintf(&buf, "function f%d() {\n", i)
		for _, g := range globals {
			fmt.Fprintf(&buf, "  global %s;\n", g)
		}
		buf.WriteString(ctx.local + "\n")
		fmt.Fprintf(&buf, "  return %s;\n}\n", test.expr)
	}
	buf.WriteString("\n")
	return buf.String()
}

type globalsWalker struct {
	globals []string
}

func (gw *globalsWalker) EnterNode(w walker.Walkable) bool {
	switch n := w.(type) {
	case *node.Root:
		return true
	case *stmt.StmtList:
		return true
	case *stmt.Expression:
		return true
	case *assign.Assign:
		name := meta.NameNodeToString(n.Variable)
		if strings.HasPrefix(name, "$") {
			gw.globals = append(gw.globals, name)
		}
		return false
	default:
		return false
	}
}

func (gw *globalsWalker) LeaveNode(walker.Walkable) {}
