package linttest_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/astutil"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/google/go-cmp/cmp"
)

// Tests in this file make it less likely that type solving will break
// without being noticed.
//
// To add a new type expr test:
// 1. Create a new test function.
// 2. For every expression that needs a type assertion use exprtype().
//
// exprtype signature is:
//   function exprtype(mixed $expr, string $expectedType)
// Where $expr is an arbitrary expression and $expectedType is a
// constant string that describes the expected type.
// Use "precise " type prefix in $expectedType if IsPrecise() is
// expected to be set.
//
// How it works:
// 1. Code being tested is indexed and then walked by exprTypeCollector.
// 2. exprTypeCollector handles exprtype() calls inside the code.
// 3. Resolved types are saved into exprTypeResult global map.
// 4. After that exprTypeWalker is executed to verify the results.

var (
	exprTypeResultMu sync.Mutex
	exprTypeResult   map[node.Node]meta.TypesMap
)

func init() {
	linter.RegisterBlockChecker(func(ctx *linter.BlockContext) linter.BlockChecker {
		return &exprTypeCollector{ctx: ctx}
	})
}

func TestExprTypePrecise2(t *testing.T) {
	code := `<?php
function test1($data) {
  $s = '123';
  exprtype($s, 'precise string');

  if ($data['key1']) {
    $s = 3.5;
    exprtype($s, 'precise float');
  } else {
    $s = 123;
    exprtype($s, 'precise int');
  }
  exprtype($s, 'precise float|int|string');

  if ($data['key2']) {
    $s = $data['x'];
    exprtype($s, 'mixed');
  }
  exprtype($s, 'float|int|mixed|string');
}

function test2($data, int $i) {
  $s = '123';

  if ($data) {
    exprtype($s, 'precise string');
  }

  if ($data) {
    $s = \UnknownClass::UNKNOWN_CONST;
  } else {
    $s = 12;
  }
  exprtype($s, 'int|string');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypePrecise(t *testing.T) {
	code := `<?php
class Foo {
  // Default value should not be considered to be precise
  // enough, since anything can be assigned later.
  public $default_int = 10;

  const STRCONST = 'abc';
}

function return_precise_int_var() {
  $local = 10;
  return $local;
}

function precise_int() { return 10; }

function typehint_int(int $i) { return $i; }

/** @param bool $b */
function repeated_info1($b) : bool { return $b; }

/** @return bool */
function repeated_info2() { return false; }

function default_bool_param($v = false) { return $v; }

/** @param bool|int $v */
function mixed_info1(int $v) {
  return $v;
}

function test() {
  $foo = new Foo();

  // TODO(quasilyte): preserve type precision when resolving
  // wrapped (lazy) type expressions.
  exprtype(precise_int(), 'int');
  exprtype(return_precise_int_var(), 'int');
  exprtype(Foo::STRCONST, 'string');

  // Cases that are debatable, but right now result in imprecise types.
  exprtype(repeated_info1(true), 'bool');
  exprtype(repeated_info2(false), 'bool');

  // Type hints are not considered to be a precise type source for now.
  // Even with strict_mode.
  exprtype(typehint_int(10), 'int');

  // Cases below should never become precise.
  exprtype($foo->default_int, 'int');
  exprtype(default_bool_param(10), 'bool');
  exprtype(mixed_info1(), 'bool|int');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeWithSpaces(t *testing.T) {
	code := `<?php
/**
 * @property $magicprop1 shape( k1: \Foo , k2 : string )
 */
class Foo {
  /** @var array<string, int> */
  public $prop1;

  /** @var $prop2 array< string, string> */
  public $prop2;

  /** @var array< string , float > $prop3 */
  public $prop3;
}

/** @param shape(a: int, b:float) $x */
function shape_param1($x) { return $x['a']; }

/** @param shape(a: int, b:float) $x */
function shape_param2($x) { return $x['b']; }

/** @param $x array{a: int, b: float} */
function array_param3($x) { return $x['a']; }

/** @param $x array{a : int, b:float} */
function array_param4($x) { return $x['b']; }

/** @return shape( x : string ) */
function shape_return1() {}

function test() {
  /** @var shape< y : int[] > $var1 */
  $var1;

  /** @var $var2 shape< z : float[] > */
  $var2;

  $foo = new Foo();

  exprtype(shape_param1($v), 'int');
  exprtype(shape_param2($v), 'float');
  exprtype(array_param3($v), 'int');
  exprtype(array_param4($v), 'float');

  exprtype($var1['y'], 'int[]');
  exprtype($var2['z'], 'float[]');

  exprtype(shape_return1()['x'], 'string');

  exprtype($foo->prop1, 'int[]');
  exprtype($foo->prop2, 'string[]');
  exprtype($foo->prop3, 'float[]');
  exprtype($foo->magicprop1['k1'], '\Foo');
  exprtype($foo->magicprop1['k2'], 'string');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeShape(t *testing.T) {
	code := `<?php
/** @param $s shape(x:int,y:float) */
function shape_self0($s) { return $s; }

/** @param $s shape(key:string) */
function shape_self1($s) { return $s; }

/** @param $s shape(nested:shape(s:string),i:integer) */
function shape_self2($s) { return $s; }

/** @param $s shape(f:double,nested:shape(i:long[])) */
function shape_self3($s) { return $s; }

/** @param shape(x?:\Foo\Bar) */
function optional_shape($s) { return $s; }

/** @param $s shape(foo:int) */
function shape_index($s) { return $s['foo']; }

/** @param $s shape(10:int,42:string) */
function shape_intkey($s) { return $s; }


/** @return shape(*) */
function shape(array $a) { return $a; }


/** @param $t tuple(int, float) */
function tuple_self0($t) { return $t; }

/** @param $t tuple(string, shape(b:bool, t:tuple(int, float))) */
function tuple_self1($t) { return $t; }

function test() {
  $s0 = shape_self0(shape(['x' => 1, 'y' => 1.5]));
  $s2 = shape_self2(shape([]));
  $s3 = shape_self3(shape([]));
  $si = shape_intkey(shape([]));
  $opt = optional_shape(shape([]));
  $t0 = tuple_self0(tuple([]));
  $t1 = tuple_self1(tuple([]));

  exprtype(shape_self0(), '\shape$exprtype.php$0$');
  exprtype(shape_self1(), '\shape$exprtype.php$1$');
  exprtype(shape_index(), 'int');

  exprtype($s0, '\shape$exprtype.php$0$');
  exprtype($s0['x'], 'int');
  exprtype($s0['y'], 'float');

  exprtype($s2['nested']['s'], 'string');
  exprtype($s2['i'], 'int');
  exprtype($s3['nested']['i'], 'int[]');
  exprtype($s3['nested']['i'][10], 'int');
  exprtype($s3['f'], 'float');

  exprtype($si[0], 'mixed');
  exprtype($si[10], 'int');
  exprtype($si[42], 'string');

  // Shapes are represented as classes and their key-type
  // info are recorded in properties map. We have a special
  // ClassShape flag to suppress field type resolving for shapes.
  exprtype($s2->i, 'mixed');
  exprtype($s0->x, 'mixed');

  // Optional keys are resolved identically.
  exprtype($opt['x'], '\Foo\Bar');

  exprtype($t0[0], 'int');
  exprtype($t0['1'], 'float');
  exprtype($t1[0], 'string');
  exprtype($t1[1]['b'], 'bool');
  exprtype($t1[1]['t'][1], 'float');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeMagicCall(t *testing.T) {
	code := `<?php
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

function test() {
  $m = new Magic();
  $m2 = new Magic2();
  $m3 = new Magic3();

  exprtype($m->magic(), '\Magic');
  exprtype($m->magic()->f2(), '\Magic');
  exprtype($m->f2()->magic(), '\Magic');
  exprtype((new Magic())->magic(), '\Magic');
  exprtype($m->notMagic(), 'int');
  exprtype($m->magic()->notMagic(), 'int');
  exprtype($m->m1()->m2()->notMagic(), 'int');

  exprtype($m2->unknown(), 'mixed');
  exprtype($m2->magicInt(), 'int');
  exprtype($m2->magicString(), 'string');
  exprtype($m2->add(1, 2), 'int');
  exprtype(Magic2::getInstance()->magicInt(), 'int');
  exprtype(Magic2::unknown(), 'mixed');

  // @method annotations should take precedence over
  // generic __call return type info.
  exprtype($m3->magicInt(), 'int');
  exprtype($m3->unknown(), '\Magic3');
  exprtype($m3->magic()->magicInt(), 'int');

  exprtype(StaticMagic::magicInt(), 'int');
  exprtype(StaticMagic::newMagic(), '\Magic');
  exprtype(StaticMagic::magic()->magic(), '\Magic');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeRef(t *testing.T) {
	code := `<?php
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

exprtype($v =& $ints[0], 'mixed');
exprtype(assign_ref_dim_fetch1(), 'mixed[]');
exprtype(assign_ref_dim_fetch2(), 'mixed[]');
exprtype(assign_ref_dim_fetch3(), 'mixed[]');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeGenerics(t *testing.T) {
	// For now, we erase most types info from the generics.

	code := `<?php
/** @return A<> */
function generic_a1() {}

/** @return A<X> */
function generic_a2() {}

/** @return A<X, Y>[] */
function generic_a3() {}

/** @return A<X, Y>|B<Z> */
function generic_a_or_b() {}

/** @return Either(int,float)|bool */
function alt_generic_intfloat() {}

exprtype(generic_a1(), '\A');
exprtype(generic_a2(), '\A');
exprtype(generic_a3(), '\A[]');
exprtype(generic_a_or_b(), '\A|\B');
exprtype(alt_generic_intfloat(), '\Either|bool');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeFixes(t *testing.T) {
	// TODO: we need to run type normalization on union types as well.
	// {`union_integer_array()`, `int|mixed[]`},
	// {`union_boolean_ints()`, `bool|int[]`},

	code := `<?php
/** @return array[] */
function array_array() {}

/** @return integer|array */
function union_integer_array() {}

/** @return boolean|[]int */
function union_boolean_ints() {}

/** @return []real */
function alias_real_arr1() {}

/** @return [][]real */
function alias_real_arr2() {}

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

exprtype(alias_double(), 'float');
exprtype(alias_real(), 'float');
exprtype(alias_integer(), 'int');
exprtype(alias_long(), 'int');
exprtype(alias_boolean(), 'bool');
exprtype(untyped_array(), 'mixed[]');
exprtype(dash(), 'mixed');
exprtype(array1(), 'int[]');
exprtype(array2(), 'int[][]');
exprtype(array_int(), 'int[]');
exprtype(array_int_string(), 'string[]');      // key type is currently ignored
exprtype(array_int_stdclass(), '\stdclass[]'); // key type is currently ignored
exprtype(array_return_string(), 'string');
exprtype(alias_real_arr1(), 'float[]');
exprtype(alias_real_arr2(), 'float[][]');
exprtype(array_array(), 'mixed[][]');
`

	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeArrayOfComplexType(t *testing.T) {
	// `(A|B)[]` is not the same as `A[]|B[]`, but it's the
	// best we can get so far.
	//
	// For nullable types, it's also not very precise.
	// `?int[]` is a nullable array, as it should be,
	// but `(?int)[]` should be interpreted differently.
	// Since we don't have real nullable types support yet,
	// we treat it identically.

	code := `<?php
/** @return (int|float)[] */
function intfloat() {}

/** @return (int|float|null)[] */
function intfloatnull() {}

/** @return ?int[] */
function nullable_int_array() {}

/** @return (?int)[] */
function array_of_nullable_ints() {}

/** @return Foo[][][] */
function array3d() {}

exprtype(intfloat(), 'float[]|int[]');
exprtype(intfloatnull(), 'float[]|int[]|null[]');
exprtype(nullable_int_array(), 'int[]|null');
exprtype(array_of_nullable_ints(), 'int[]|null');
exprtype(array3d(), '\Foo[][][]');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeVoid(t *testing.T) {
	code := `<?php
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

function test() {
  $foo = new Foo();

  exprtype(void_func1(), 'void');
  exprtype(void_func2(), 'void');
  exprtype(void_func3(), 'void');
  exprtype($foo->voidMeth1(), 'void');
  exprtype($foo->voidMeth2(), 'void');
  exprtype($foo->voidMeth3(), 'void');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeArrayAccess(t *testing.T) {
	code := `<?php
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

function test() {
  $ints = new Ints(); $self = new Self();

  exprtype($ints[0], 'int');
  exprtype(getInts()[0], 'int');
  exprtype($self[0], '\Self');
  exprtype($self[0][1], '\Self');
  exprtype($self[0][1]->offsetGet(2), '\Self');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeAnnotatedProperty(t *testing.T) {
	code := `<?php
/**
 * @property int $int optional description
 */
class Foo {
  /***/
  public function getInt() {
    return $this->int;
  }
}

function test() {
  $x = new Foo();
  exprtype($x->int, 'int');
  exprtype($x->getInt(), 'int');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeScopeNoreplace(t *testing.T) {
	// These tests cover special NoReplace flag of the meta.ScopeVar.

	code := `<?php
/** @param string $v */
function phpdoc_param($v) {
  $v = 10;
  return $v;
}

function phpdoc_localvar() {
  /** @var string $x */
  $x = '123';
  $x = 10;
  return $x;
}

function localvar() {
  $x = '123';
  $x = 10;
  return $x;
}

exprtype(phpdoc_param($v), 'int');
exprtype(phpdoc_localvar(), 'int|string');
exprtype(localvar(), 'int');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeMalformedPhpdoc(t *testing.T) {
	code := `<?php
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

exprtype(return_mixed(0), 'mixed');
exprtype(return_int(0), 'int');
exprtype(return_int2(0), 'int');
exprtype(return_int3(0), 'int');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeMagicGet(t *testing.T) {
	code := `
class Ints {
  public function __get($k) { return 0; }
}
class Chain {
  public function __get($k) { return $this; }
}

function test() {
  $ints = new Ints();
  $chain = new Chain();

  exprtype((new Ints)->a, 'int');
  exprtype($ints->a, 'int');
  exprtype($ints->b, 'int');
  exprtype((new Chain)->chain, '\Chain');
  exprtype($chain->chain, '\Chain');
  exprtype($chain->chain->chain, '\Chain');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeHint(t *testing.T) {
	code := `
function array_hint(array $x) { return $x; }
function callable_hint(callable $x) { return $x; }

function integer_hint(integer $x) { return $x; }
function boolean_hint(boolean $x) { return $x; }
function real_hint(real $x) { return $x; }
function double_hint(double $x) { return $x; }

function integer_hint2() : integer {}
function boolean_hint2() : boolean {}
function real_hint2() : real {}
function double_hint2() : double {}

exprtype(array_hint(), 'mixed[]');
exprtype(callable_hint(), 'callable');

exprtype(integer_hint(), '\integer');
exprtype(boolean_hint(), '\boolean');
exprtype(real_hint(), '\real');
exprtype(double_hint(), '\double');
exprtype(integer_hint2(), '\integer');
exprtype(boolean_hint2(), '\boolean');
exprtype(real_hint2(), '\real');
exprtype(double_hint2(), '\double');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeNullable(t *testing.T) {
	code := `
class A {
  /** @var ?B */
  public $b;
}
class B {
  public $c;
}

/**
 * @return ?int
 */
function nullable_int($cond) {
  if ($cond) {
    return 4;
  }
  return null;
}


/**
 * @return ?int[]
 */
function nullable_array($cond) {
  if ($cond) {
    return [1];
  }
  return null;
}

function nullable_string($cond) : ?string {
  if ($cond) {
    return '123';
  }
  return null;
}

function test() {
  /** @var ?int $int */
  $int = null;

  /** @var ?int|?string $foo */
  $foo = null;

  $a = new A();

  exprtype($int, 'int|null');
  exprtype($foo, 'int|string|null');
  exprtype($a->b, '\B|null');
  exprtype(nullable_int(1), 'int|null');
  exprtype(nullable_string(0), 'string|null');
  exprtype(nullable_array(0), 'int[]|null');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeLateStaticBinding(t *testing.T) {
	code := `
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

function test() {
  $b = new Base();
  $d = new Derived();
  $dd = new DerivedDerived();

  exprtype(getBase(), '\Base');
  exprtype(getDerived(), '\Base|\Derived');
  exprtype(getBase2(), '\Base');
  exprtype(getDerived2(), '\Base|\Derived');
  exprtype(getBase2()->getStatic()->getStatic(), '\Base');
  exprtype(getDerived2()->getStatic()->getStatic(), '\Base|\Derived');
  exprtype(eitherDerived(), '\Derived|\DerivedDerived');
  exprtype(eitherDerived()->getStatic(), '\Base|\Derived|\DerivedDerived');

  exprtype(Base::staticNewStatic(), '\Base');
  exprtype(Base::staticNewStatic()->staticNewStatic(), '\Base');
  exprtype(Derived::staticNewStatic(), '\Derived');
  exprtype(Derived::staticNewStatic()->staticNewStatic(), '\Derived');
  exprtype(DerivedDerived::staticNewStatic(), '\DerivedDerived');
  exprtype(DerivedDerived::staticNewStatic()->staticNewStatic(), '\DerivedDerived');

  exprtype($b->newStatic(), '\Base');
  exprtype($d->newStatic(), '\Derived');
  exprtype($dd->newStatic(), '\DerivedDerived');

  exprtype($b->getStatic(), '\Base');
  exprtype($b->getStatic()->getStatic(), '\Base');
  exprtype($b->getStaticArray(), '\Base[]');
  exprtype($b->getStaticArray()[0], '\Base');
  exprtype($b->getStaticArrayArray(), '\Base[][]');
  exprtype($b->getStaticArrayArray()[0][0], '\Base');

  exprtype($d->getStatic(), '\Base|\Derived');
  exprtype($d->getStatic()->getStatic(), '\Base|\Derived');
  exprtype($d->getStaticArray(), '\Derived[]');
  exprtype($d->getStaticArray()[0], '\Derived');
  exprtype($d->getStaticArrayArray(), '\Derived[][]');
  exprtype($d->getStaticArrayArray()[0][0], '\Derived');

  exprtype($dd->getStatic(), '\Base|\DerivedDerived');
  exprtype($dd->getStatic()->getStatic(), '\Base|\DerivedDerived');
  exprtype($dd->getStaticArray(), '\DerivedDerived[]');
  exprtype($dd->getStaticArray()[0], '\DerivedDerived');
  exprtype($dd->getStaticArrayArray(), '\DerivedDerived[][]');
  exprtype($dd->getStaticArrayArray()[0][0], '\DerivedDerived');

  exprtype($b->initAndReturnOther1(), '\Base');
  exprtype($b->initAndReturnOther2(), '\Base');

  exprtype((new Base())->getStatic(), '\Base');
  exprtype((new Derived())->getStatic(), '\Base|\Derived');

  exprtype($d->derivedGetStatic(), '\Derived');
  exprtype($d->derivedNewStatic(), '\Derived');
  exprtype($dd->derivedGetStatic(), '\Derived|\DerivedDerived');
  exprtype($dd->derivedNewStatic(), '\DerivedDerived');

  exprtype($d->getStatic(), '\Base|\Derived');
  exprtype($d->getStatic()->getStatic(), '\Base|\Derived');
  exprtype($dd->getStatic(), '\Base|\DerivedDerived');
  exprtype($dd->getStatic()->getStatic(), '\Base|\DerivedDerived');

  exprtype($d->getStaticForOverride1(), 'null|\Derived');
  exprtype($d->getStaticForOverride2(), '\Derived');
  exprtype($d->getStaticForOverride3(), '\Derived');
  exprtype($dd->getStaticForOverride1(), 'null|\DerivedDerived');
  exprtype($dd->getStaticForOverride2(), '\Derived'); // Since $this works like 'self'
  exprtype($dd->getStaticForOverride3(), '\Derived|\DerivedDerived');

  exprtype($dd->asParent(), '\Derived|\DerivedDerived');
  exprtype($dd->asParent()->newStatic(), '\Derived|\DerivedDerived');
  exprtype($dd->asParent()->asParent(), '\Derived|\DerivedDerived');

  // Resolving of '$this' (which should be identical to 'static').
  exprtype($b->getThis(), '\Base');
  exprtype($d->getThis(), '\Base|\Derived');
  exprtype($b->getThis()->getThis(), '\Base');
  exprtype($d->getThis()->getThis(), '\Base|\Derived');

  // TODO: resolve $this without @return hint into 'static' as well?
  exprtype($b->getThisNoHint(), '\Base');
  exprtype($d->getThisNoHint(), '\Base');
  exprtype($dd->getThisNoHint(), '\Base');
}
`

	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeSimple(t *testing.T) {
	code := `<?php
class Foo {}

function define($name, $value) {}
define('true', (bool)1);
define('false', (bool)0);
$int = 10;
$float = 20.5;
$string = "123";

function empty_array() { $x = []; return $x; }

function test() {
  global $int;
  global $float;
  global $string;

  exprtype(true, 'precise bool');
  exprtype(false, 'precise bool');
  exprtype((bool)1, 'precise bool');
  exprtype((boolean)1, 'precise bool');

  exprtype(1, 'precise int');
  exprtype((int)1.5, 'precise int');
  exprtype((integer)1.5, 'precise int');

  exprtype(1.21, 'precise float');
  exprtype((float)1, 'precise float');
  exprtype((real)1, 'precise float');
  exprtype((double)1, 'precise float');

  exprtype("", 'precise string');
  exprtype((string)1, 'precise string');

  exprtype([], 'mixed[]');
  exprtype([1, 'a', 4.5], 'mixed[]');

  exprtype(1+5<<2, 'precise int');

  exprtype(-1, 'int');
  exprtype(-1.4, 'float');
  exprtype(+1, 'int');
  exprtype(+1.4, 'float');

  exprtype(~$int, 'int');
  exprtype(~'dsds', 'string');

  exprtype($int & $int, 'int');

  exprtype($float & $int, 'int');
  exprtype($int & $float, 'int');
  exprtype(4.5 & 1.4, 'int');
  exprtype("abc" & "foo", 'string');
  exprtype($int | $int, 'int');
  exprtype(4.5 | 1.4, 'int');
  exprtype("abc" | "foo", 'string');
  exprtype($int ^ $int, 'int');
  exprtype(4.5 ^ 1.4, 'int');
  exprtype("abc" ^ "foo", 'string');

  exprtype($int, 'int');
  exprtype($float, 'float');
  exprtype($string, 'string');

  exprtype(define('foo', 0 == 0), 'void');
  exprtype(empty_array(), 'mixed[]');

  exprtype(new Foo(), 'precise \Foo');
  exprtype(clone (new Foo()), 'precise \Foo');

  exprtype(1 > 4, 'precise bool');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeKeyword(t *testing.T) {
	code := `<?php
/** @return resource */
function f_resource() {}
exprtype(f_resource(), 'resource');

/** @return true */
function f_true() {}
exprtype(f_true(), 'true');

/** @return false */
function f_false() {}
exprtype(f_false(), 'false');

/** @return iterable */
function f_iterable() {}
exprtype(f_iterable(), 'iterable');

/** @return (resource[]) */
function f_resource2() {}
exprtype(f_resource2(), 'resource[]');

/** @return (true[]) */
function f_true2() {}
exprtype(f_true2(), 'true[]');

/** @return (false[]) */
function f_false2() {}
exprtype(f_false2(), 'false[]');

/** @return (iterable[]) */
function f_iterable2() {}
exprtype(f_iterable2(), 'iterable[]');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeArray(t *testing.T) {
	code := `
function test() {
  $int = 10;
  $ints = [1, 2];

  exprtype([], 'mixed[]'); // Should never be "empty_array" after resolving
  exprtype([[]], 'mixed[]');
  exprtype([1, 2], 'int[]');
  exprtype([1.4, 3.5], 'float[]');
  exprtype(["1", "5"], 'string[]');
  exprtype(["k1" => 123, "k2" => 345], 'int[]');
  exprtype([0 => "a", 1 => "b"], 'string[]');

  exprtype([$int, $int], 'mixed[]'); // TODO: could be int[]
  exprtype($ints[0], 'int');
  exprtype(["11"][0], 'string');
  exprtype([1.4][0], 'float');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeMulti(t *testing.T) {
	code := `
/** @return mixed */
function get_mixed($x) { return $x; }

$cond = "true";
$int_or_float = 10;
if ($cond) {
  $int_or_float = 10.5;
}

function test() {
  global $int_or_float;
  global $cond;

  /** @var bool|int $bool_or_int */
  $bool_or_int = 10;

  exprtype($cond ? 1 : 2, 'precise int');
  exprtype($int_or_float, 'int|float');
  exprtype($cond ? 10 : '123', 'precise int|string');
  exprtype($cond ? ($int_or_float ? 10 : 10.4) : (bool)1, 'precise int|float|bool');
  exprtype($bool_or_int, 'bool|int');
  exprtype($cond ? 10 : get_mixed(1), 'int|mixed');
  exprtype($cond ? get_mixed(1) : 10, 'int|mixed');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeOps(t *testing.T) {
	code := `<?php
$global_int = 10;
$global_float = 20.5;

function test() {
  global $global_int;
  global $global_float;

  $int = 10;
  $float = 20.5;
  $string = "123";
  $bool = (bool)1;

  exprtype(1 + $int, 'int');
  exprtype($int + 1, 'int');
  exprtype(1 + (int)$float, 'int');
  exprtype(1 + $global_int, 'float'); // TODO: should be int?
  exprtype(1 + $float, 'float');
  exprtype($int . $float, 'precise string');
  exprtype($int && $float, 'precise bool');
  exprtype($int || 1, 'precise bool');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeProperty(t *testing.T) {
	code := `<?php

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
  /** @var double */
  public $x;
  /** @var float */
  public $y;
}

function test() {
  $point = new Point();
  $magic = new Magic();

  exprtype($point->x, 'float');
  exprtype($point->y, 'float');
  exprtype(Gopher::$name, 'string');
  exprtype(Gopher::POWER, 'int');
  exprtype($magic->int, 'int');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeFunction(t *testing.T) {
	code := `<?php
function define($name, $value) {}
define('null', 0);

class Foo {}

function mixed_array($x) {
  return [$x, 1, 2];
}
exprtype(mixed_array(0), 'mixed[]');
exprtype(mixed_array(0)[1], 'mixed');

function mixed_or_ints1($x) {
  if ($x) {
    return mixed_array($x);
  }
  return [0, 0];
}
exprtype(mixed_or_ints1(0), 'int[]|mixed[]');
exprtype(mixed_or_ints1(0)[1], 'int|mixed');

function mixed_or_ints2($x) {
  $a = array(0, 0);
  if ($x) {
    $a = mixed_array($x);
  }
  return $a;
}
exprtype(mixed_or_ints2(0), 'int[]|mixed[]');
exprtype(mixed_or_ints2(0)[1], 'int|mixed');

function recur1($cond) {
  if ($cond) { return 0; }
  return recur2($cond);
}
exprtype(recur1(true), 'int|string');

function recur2($cond) {
  if ($cond) { return ""; }
  return recur1($cond);
}
exprtype(recur2(true), 'int|string');

function recur3() { return recur4(); }
function recur4() { return recur5(); }
function recur5() { return recur3(); }

exprtype(recur3(), 'mixed');
exprtype(recur4(), 'mixed');
exprtype(recur5(), 'mixed');

function bare_ret1($cond) {
  if ($cond) { return; }
  return 10;
}
exprtype(bare_ret1(false), 'int|null');

function bare_ret2($cond) {
  if ($cond) { return 10; }
  return;
}
exprtype(bare_ret2(false), 'int|null');

function untyped_param($x) { return $x; }
exprtype(untyped_param(0), 'mixed');

function undefined_type1() {
  $x = unknown_func();
  return $x;
}
exprtype(undefined_type1(), 'mixed');

function undefined_type2() {
  return $x;
}
exprtype(undefined_type2(), 'mixed');

function foreach1($xs) {
  foreach ($xs as $_) {
    return 10;
  }
  return "";
}
exprtype(foreach1([]), 'int|string');

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
exprtype(foreach2([]), 'int|string');

function throw1($cond) {
  if ($cond) {
    return 10;
  }
  throw new Exception();
}
exprtype(throw1(true), 'int');

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
exprtype(throw2(true), 'bool|int');

function get_ints() {
  $a = []; // "empty_array"
  $a[0] = 1;
  $a[1] = 2;
  return $a; // Should be resolved to just int[]
}
exprtype(get_ints(), 'int[]');

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
exprtype(switch1(true), 'bool|int|string');

function switch2($v) {
  switch ($v) {
  case 10:
    return 10;
  case 20:
    return "";
  }
  return false;
}
exprtype(switch2(true), 'bool|int|string');

function switch3($v) {
  switch ($v) {
  default:
    return "";
  }
  return false;
}
exprtype(switch3(true), 'bool|string');

function ifelse1($cond) {
  if ($cond) {
    return 10;
  } else if ($cond+1) {
    return "";
  } else {
    return false;
  }
}
exprtype(ifelse1(true), 'bool|int|string');

function ifelse2($cond) {
  if ($cond) {
    return 10;
  } elseif ($cond+1) {
    return "";
  } else {
    return false;
  }
}
exprtype(ifelse2(true), 'bool|int|string');

function ifelse3($cond) {
  if ($cond) {
    return 10;
  } elseif ($cond+1) {
    return "";
  }
  return false;
}
exprtype(ifelse3(true), 'bool|int|string');

function try_catch1() {
  try {
    return 10;
  } catch (Exception $_) {
    return "";
  }
  return false;
}
exprtype(try_catch1(), 'bool|int|string');

function try_finally1() {
  try {
    return 10;
  } finally {
    return "";
  }
  return false;
}
exprtype(try_finally1(), 'bool|int|string');

/** @return float[] */
function get_floats() { return []; }
exprtype(get_floats(), 'float[]');

function get_array() { return []; }
exprtype(get_array(), 'mixed[]');

/** @return array */
function get_array_or_null() { return null; }
exprtype(get_array_or_null(), 'mixed[]|null');

/** @return null */
function get_null_or_array() { return []; }
exprtype(get_null_or_array(), 'mixed[]|null');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeMethod(t *testing.T) {
	code := `<?php
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
}

namespace {
  function f() {
    $test = new \NS\Test();
    $derived = new Derived();

    exprtype(\NS\Test::instance(), '\NS\Test');
    exprtype(\NS\Test::instance2(), '\NS\Test');
    exprtype($test->getInt(), 'int');
    exprtype($test->getInts(), 'int[]');
    exprtype($test->getThis()->getThis()->getInt(), 'int');
    exprtype(new \NS\Test(), 'precise \NS\Test');
  }
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeInterface(t *testing.T) {
	code := `<?php
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
}

exprtype(new Foo(), 'precise \Foo');

function f() {
  $foo = new Foo();
  exprtype($foo, 'precise \Foo');
  exprtype($foo->getThis(), '\Foo');
  exprtype($foo->acceptThis($foo), '\TestInterface');
  exprtype($foo->acceptThis($foo)->acceptThis($foo), '\TestInterface');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeOverride(t *testing.T) {
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
	code := `<?php
$ints = [1, 2, 3];
exprtype($ints, 'int[]');

function returns_array_shift() {
  global $ints;
  return array_shift($ints);
}
exprtype(returns_array_shift(), 'int');
exprtype(array_shift($ints), 'int|mixed');

function returns_array_slice() {
  global $ints;
  return array_slice($ints, 0, 2);
}
exprtype(returns_array_slice(), 'int[]');

exprtype(array_map(function($x) { return $x; }, $ints), 'mixed[]');
`
	runExprTypeTest(t, &exprTypeTestParams{stubs: stubs, code: code})
}

func runExprTypeTest(t *testing.T, params *exprTypeTestParams) {
	meta.ResetInfo()
	if params.stubs != "" {
		linter.InitStubs(func(ch chan linter.FileInfo) {
			ch <- linter.FileInfo{
				Filename: "stubs.php",
				Contents: []byte(params.stubs),
			}
		})
	}
	linttest.ParseTestFile(t, "exprtype.php", params.code)

	meta.SetIndexingComplete(true)

	// Reset results map and run expr type collector.
	exprTypeResult = map[node.Node]meta.TypesMap{}
	root, _ := linttest.ParseTestFile(t, "exprtype.php", params.code)

	// Check that collected types are identical to the expected types.
	// We need the second walker to pass *testing.T parameter to
	// the walker that does the comparison.
	walker := exprTypeWalker{t: t}
	root.Walk(&walker)
}

type testTypesMap struct {
	Precise bool
	Types   string
}

func makeType(typ string) testTypesMap {
	if typ == "" {
		return testTypesMap{}
	}

	precise := strings.HasPrefix(typ, "precise ")
	if precise {
		typ = strings.TrimPrefix(typ, "precise ")
	}

	return testTypesMap{Precise: precise, Types: typ}
}

type exprTypeTestParams struct {
	code  string
	stubs string
}

type exprTypeWalker struct {
	t *testing.T
}

func (w *exprTypeWalker) LeaveNode(n walker.Walkable) {}

func (w *exprTypeWalker) EnterNode(n walker.Walkable) bool {
	call, ok := n.(*expr.FunctionCall)
	if ok && meta.NameNodeEquals(call.Function, `exprtype`) {
		checkedExpr := call.ArgumentList.Arguments[0].(*node.Argument).Expr
		expectedType := call.ArgumentList.Arguments[1].(*node.Argument).Expr.(*scalar.String).Value
		actualType, ok := exprTypeResult[checkedExpr]
		if !ok {
			w.t.Fatalf("no type found for %s expression", astutil.FmtNode(checkedExpr))
		}
		want := makeType(expectedType[len(`"`) : len(expectedType)-len(`"`)])
		have := testTypesMap{
			Types:   actualType.String(),
			Precise: actualType.IsPrecise(),
		}
		if diff := cmp.Diff(have, want); diff != "" {
			line := checkedExpr.GetPosition().StartLine
			w.t.Errorf("line %d: type mismatch for %s (-have +want):\n%s",
				line, astutil.FmtNode(checkedExpr), diff)
		}
		return false
	}

	return true
}

type exprTypeCollector struct {
	ctx *linter.BlockContext
	linter.BlockCheckerDefaults
}

func (c *exprTypeCollector) AfterEnterNode(n walker.Walkable) {
	if !meta.IsIndexingComplete() {
		return
	}

	call, ok := n.(*expr.FunctionCall)
	if !ok || !meta.NameNodeEquals(call.Function, `exprtype`) {
		return
	}
	checkedExpr := call.ArgumentList.Arguments[0].(*node.Argument).Expr

	// We need to clone a types map because if it belongs to a var
	// or some other symbol those type can be volatile we'll get
	// unexpected results.
	typ := c.ctx.ExprType(checkedExpr).Clone()

	exprTypeResultMu.Lock()
	exprTypeResult[checkedExpr] = typ
	exprTypeResultMu.Unlock()
}
