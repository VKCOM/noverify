package exprtype_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/VKCOM/noverify/src/utils"
	"github.com/google/go-cmp/cmp"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/types"
	"github.com/VKCOM/noverify/src/workspace"
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
	exprTypeResult   map[ir.Node]types.Map
)

func TestExprTypeListOverArray(t *testing.T) {
	code := `<?php
/**
 * @param int[] $xs
 */
function ints($xs) {
  list ($a, $b) = $xs;
  exprtype($a, 'int');
  exprtype($b, 'int');
}

/**
 * @param string[]|false $xs
 */
function strings_or_false($xs) {
  list ($a, $b) = $xs;
  exprtype($a, 'string');
  exprtype($b, 'string');
}

/**
 * @param null|\Foo[]|int $xs
 */
function null_or_foos_or_int($xs) {
  list ($a, $b) = $xs;
  exprtype($a, '\Foo');
  exprtype($b, '\Foo');
}

/**
 * @param int[]|string[] $xs
 */
function ints_or_strings($xs) {
  list ($a, $b) = $xs;
  exprtype($a, 'int|string');
  exprtype($b, 'int|string');
}

/**
 * @param mixed[]|string[]|int[]|false $xs
 */
function mixeds_or_strings_or_ints_or_false($xs) {
  list ($a, $b) = $xs;
  exprtype($a, 'int|mixed|string');
  exprtype($b, 'int|mixed|string');
}

/**
 * @param int|float|null $xs
 */
function not_an_array($xs) {
  list ($a, $b) = $xs;
  exprtype($a, 'unknown_from_list');
  exprtype($b, 'unknown_from_list');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeCatch(t *testing.T) {
	code := `<?php
try {
} catch (Foo $x) {
  exprtype($x, '\Foo');
} catch (Exception $e) {
  exprtype($e, '\Exception');
} catch (A|B $ab) {
  exprtype($ab, '\A|\B');
} catch (\A\B\C $abc) {
  exprtype($abc, '\A\B\C');
  exprtype($ab, 'undefined');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeVariadicParam(t *testing.T) {
	code := `<?php
function scalar_int(int ...$xs) {
  exprtype($xs, 'int[]');
}

function no_typehint(...$xs) {
  exprtype($xs, 'mixed'); // TODO: mixed[]?
}

function foo_array(Foo ...$xs) {
  exprtype($xs, '\Foo[]');
}

function mixed_array2(array ...$xs) {
  exprtype($xs, 'mixed[][]');
}

/** @param $xs Foo */
function scalar_int(int ...$xs) {
  exprtype($xs, '\Foo|int[]');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeForeachKey(t *testing.T) {
	code := `<?php
$xs = [[1], [2]];

foreach ($xs as $k => $ys) {
  exprtype($k, 'int|string');

  foreach ($xs as $k2 => $_) {
    exprtype($k2, 'int|string');
    $k2 = 10;
    exprtype($k2, 'precise int');
  }

  exprtype($k, 'int|string');

  $v = $xs ? $k : [1];
  exprtype($v, 'int|int[]|string');
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeRecursiveType1(t *testing.T) {
	code := `<?php
class Feed {
  /** @var FeedItem[] */
  public $items;
}

class FeedItem {
  /**
   * @var FeedItem[]
   */
  public $stories;

  public $title = '';
}

exprtype((new FeedItem())->stories, '\FeedItem[]');
exprtype((new Feed())->items[0], '\FeedItem');

$feed = new Feed();
exprtype($feed->items[0]->stories, '\FeedItem[]');

function test(Feed $feed) {
  exprtype($feed->items, '\FeedItem[]');

  foreach ($feed->items as $item) {
    exprtype($item, '\FeedItem');
    exprtype($item->stories, '\FeedItem[]');

    foreach ($item->stories as $story) {
      exprtype($story, '\FeedItem');
      $_ = $story->title;
    }
  }
}
`

	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeRecursiveType2(t *testing.T) {
	code := `<?php
class MyList {
  /** @var MyList */
  public $tail;

  /** @return MyList */
  public function getTail() { return $this->tail; }
}

/** @return MyList[][] */
function newList() { return [[new MyList()]]; }

$l = new MyList();

exprtype($l->tail, '\MyList');
exprtype($l->tail->tail, '\MyList');
exprtype($l->tail->tail->tail, '\MyList');

exprtype($l->getTail(), '\MyList');
exprtype($l->getTail()->getTail(), '\MyList');
exprtype($l->getTail()->getTail()->getTail(), '\MyList');

exprtype((newList())[0][0]->getTail(), '\MyList');
exprtype((newList())[0][0]->getTail()->getTail(), '\MyList');
exprtype((newList())[0][0]->getTail()->getTail()->getTail(), '\MyList');

class A {
  /** @var B */
  public $b;
}
class B {
  /** @var A */
  public $a;
}

$loop = new A();
exprtype($loop->b, '\B');
exprtype($loop->b->a, '\A');
exprtype($loop->b->a->b, '\B');
exprtype($loop->b->a->b->a, '\A');
`

	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeTraitSelfStatic1(t *testing.T) {
	// Tests for WStaticMethodCall.
	code := `<?php
trait NewSelf {
  public static function instance() { return new self(); }
}

class FooSelf { use NewSelf; }
class BarSelf extends FooSelf {}
class BazSelf extends BarSelf { use NewSelf; }

exprtype(FooSelf::instance(), '\FooSelf');
exprtype(BarSelf::instance(), '\FooSelf');
exprtype(BazSelf::instance(), '\BazSelf');

trait NewStatic {
  public static function instance() { return new static(); }
}

class FooStatic { use NewStatic; }
class BarStatic extends FooStatic {}
class BazStatic extends BarStatic { use NewStatic; }

exprtype(FooStatic::instance(), '\FooStatic');
exprtype(BarStatic::instance(), '\BarStatic');
exprtype(BazStatic::instance(), '\BazStatic');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeTraitSelfStatic2(t *testing.T) {
	// Tests for WStaticPropertyFetch.
	code := `<?php
trait NewSelf {
  /** @var self */
  public static $v = null;
}

class FooSelf {
  use NewSelf;
  private static function f() {
    exprtype(self::$v, '\FooSelf|null');
  }
}

class BarSelf extends FooSelf {
  private static function f() {
    exprtype(self::$v, '\FooSelf|null');
  }
}

class BazSelf {
  use NewSelf;
  private static function f() {
    exprtype(self::$v, '\BazSelf|null');
  }
}

exprtype(FooSelf::$v, '\FooSelf|null');
exprtype(BarSelf::$v, '\FooSelf|null');
exprtype(BazSelf::$v, '\BazSelf|null');

trait NewStatic {
  /** @var static */
  public static $v = null;
}

class FooStatic {
  use NewStatic;
  private static function f() {
    exprtype(self::$v, '\FooStatic|null');
  }
}

class BarStatic extends FooStatic {
  private static function f() {
    exprtype(self::$v, '\BarStatic|null');
  }
}

class BazStatic {
  use NewStatic;
  private static function f() {
    exprtype(self::$v, '\BazStatic|null');
  }
}

exprtype(FooStatic::$v, '\FooStatic|null');
exprtype(BarStatic::$v, '\BarStatic|null');
exprtype(BazStatic::$v, '\BazStatic|null');
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestExprTypeIssue497(t *testing.T) {
	code := `<?php
/**
 * @param shape(a:int) $x
 *
 * @return T<int>
 */
function f($x) {
  exprtype($x, '\shape$a:int$');
  return [$v];
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
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

  exprtype(shape_self0(), '\shape$x:int,y:float$');
  exprtype(shape_self1(), '\shape$key:string$');
  exprtype(shape_index(), 'int');

  exprtype($s0, '\shape$x:int,y:float$');
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
exprtype(assign_ref_dim_fetch1(), 'int[][]');
exprtype(assign_ref_dim_fetch2(), 'int[]');
exprtype(assign_ref_dim_fetch3(), 'int[][]');
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

function bare_ret3($cond) {
  if ($cond == 1) { return 10; }
  if ($cond == 2) { return ""; }
  return;
}
exprtype(bare_ret3(10), 'int|null|string');

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
exprtype(get_array_or_null(), 'mixed[]');

/** @return null */
function get_null_or_array() { return []; }
exprtype(get_null_or_array(), 'null');
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

func TestPostfixPrefixIncDec(t *testing.T) {
	code := `<?php

$a = 100;

exprtype($a++, "int");
exprtype($a--, "int");
exprtype(++$a, "int");
exprtype(--$a, "int");


$a = 100.5;

exprtype($a++, "float");
exprtype($a--, "float");
exprtype(++$a, "float");
exprtype(--$a, "float");


$a = "100";

exprtype($a++, "float");
exprtype($a--, "float");
exprtype(++$a, "float");
exprtype(--$a, "float");


if ($a == 100) {
	$a = 56.5
} else {
	$a = 56
}

exprtype($a++, "float");
exprtype($a--, "float");
exprtype(++$a, "float");
exprtype(--$a, "float");


class Foo {};

$a = new Foo();

exprtype($a++, "float");
exprtype($a--, "float");
exprtype(++$a, "float");
exprtype(--$a, "float");

`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestTypesListOverTuple(t *testing.T) {
	code := `<?php
class Boo {}

/**
 * @return \tuple(int, \Boo, int, string)
 */
function foo() {
   return [5, new Boo(), 10, "gas"];
}

$tuple = foo();

[$i, $bo, $j, $str] = $tuple; // With short syntax

exprtype($i, "int");
exprtype($bo, "\Boo");
exprtype($j, "int");
exprtype($str, "string");


list($old_i, $old_bo, $old_j, $old_str) = $tuple; // With old syntax

exprtype($old_i, "int");
exprtype($old_bo, "\Boo");
exprtype($old_j, "int");
exprtype($old_str, "string");


class Bar {
   /**
    * @return \tuple(int, \Boo, int, string)
    */
   static function foo() {
      return [5, new Boo(), 10, "gas"];
   }
}

$class_static_tuple = Bar::foo();

[$class_static_i, $class_static_bo, $class_static_j, $class_static_str] = $class_static_tuple;

exprtype($class_static_i, "int");
exprtype($class_static_bo, "\Boo");
exprtype($class_static_j, "int");
exprtype($class_static_str, "string");


[$function_i, $function_bo, $function_j, $function_str] = Bar::foo();

exprtype($function_i, "int");
exprtype($function_bo, "\Boo");
exprtype($function_j, "int");
exprtype($function_str, "string");

$tuple_int = 10;
[$tuple_int_i, $tuple_int_foo, $tuple_int_j, $tuple_int_str] = $tuple_int;

exprtype($tuple_int_i, "unknown_from_list");
exprtype($tuple_int_foo, "unknown_from_list");
exprtype($tuple_int_j, "unknown_from_list");
exprtype($tuple_int_str, "unknown_from_list");
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestTypesShapeOverList(t *testing.T) {
	code := `<?php
class Foo {}

/** @return shape(key:\Foo, key2:float, key3:int) */
function asShapeWithStringKey() { return []; }

/** @return shape(0:\Foo, 2:float, 1:int) */
function asShapeWithIntKey() { return []; }

/** @return shape(0:\Foo, 4:float, 2:int) */
function asShapeWithSomeIntKey() { return []; }

function foo() {
  // simple
  ["key" => $a1, "key2" => $b1, "key3" => $c1] = asShapeWithStringKey();

  exprtype($a1, "\Foo");
  exprtype($b1, "float");
  exprtype($c1, "int");


  // mixed positions
  ["key" => $a2, "key3" => $c2, "key2" => $b2] = asShapeWithStringKey();

  exprtype($a2, "\Foo");
  exprtype($b2, "float");
  exprtype($c2, "int");


  // without keys and shape with string key
  [$a3, $c3, $b3] = asShapeWithStringKey();

  exprtype($a3, "unknown_from_list");
  exprtype($b3, "unknown_from_list");
  exprtype($c3, "unknown_from_list");


  // without keys and shape with int key
  [$a4, $b4, $c4] = asShapeWithIntKey();

  exprtype($a4, "\Foo");
  exprtype($b4, "int");
  exprtype($c4, "float");


  // without keys and shape with some int key
  [$a5, $b5, $c5, $d5, $e5] = asShapeWithSomeIntKey();

  exprtype($a5, "\Foo");
  exprtype($b5, "unknown_from_list");
  exprtype($c5, "int");
  exprtype($d5, "unknown_from_list");
  exprtype($e5, "float");


  // with keys and shape with some int key
  [0 => $a6, 2 => $b6, 4 => $c6] = asShapeWithSomeIntKey();

  exprtype($a6, "\Foo");
  exprtype($b6, "int");
  exprtype($c6, "float");


  // with old style and keys and shape with some int key
  list(0 => $a7, 2 => $b7, 4 => $c7) = asShapeWithSomeIntKey();

  exprtype($a7, "\Foo");
  exprtype($b7, "int");
  exprtype($c7, "float");


  // with old style and without keys and shape with string key
  list($a8, $c8, $b8) = asShapeWithStringKey();

  exprtype($a8, "unknown_from_list");
  exprtype($b8, "unknown_from_list");
  exprtype($c8, "unknown_from_list");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestArrayTypes(t *testing.T) {
	code := `<?php
class Foo {}

/** @return Foo */
function f() {}
/** @return float */
function return_float() {}
/** @return int */
function return_int() {}
/** @return string */
function return_string() {}


$one_dimensional = [new Foo(), new Foo()];
exprtype($one_dimensional, "\Foo[]");


$two_dimensional = [[new Foo(), new Foo()],[new Foo(), new Foo()]];
exprtype($two_dimensional, "\Foo[][]");


$three_dimensional = [[[new Foo(), new Foo()],[new Foo(), new Foo()]],[[new Foo(), new Foo()],[new Foo(), new Foo()]]];
exprtype($three_dimensional, "\Foo[][][]");


$a = [10, 20, 30];
exprtype($a, "int[]");
$a = [return_int(), return_int(), return_int()];
exprtype($a, "int[]");
// but
$a = [return_int(), 1];
exprtype($a, "mixed[]");


$a = [10.5, 20.5, 30.5];
exprtype($a, "float[]");
$a = [return_float(), return_float(), return_float()];
exprtype($a, "float[]");
// but
$a = [return_float(), 12.5];
exprtype($a, "mixed[]");


$a = ["Hello", "World", "!"];
exprtype($a, "string[]");
$a = [return_string(), return_string(), return_string()];
exprtype($a, "string[]");
// but
$a = [return_string(), "World!"];
exprtype($a, "mixed[]");


$a = [f(), f()];
exprtype($a, "\Foo[]");
// but
$a = [f(), new Foo()];
exprtype($a, "mixed[]");


$a = [f(), g()];
exprtype($a, "mixed[]");


$a = [];
exprtype($a, "mixed[]");
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestArrayTypesWithUnpack(t *testing.T) {
	code := `<?php
class Foo {}

/** @return Boo[] */
function returnBooArray() {}

function f() {
  // with simple type
  $a = [1,2,3];
  $b = [0, ...$a];
  exprtype($b, "int[]");


  // with class type
  $a1 = [new Foo(), new Foo()];
  $b1 = [new Foo(), ...$a1];
  exprtype($b1, "\Foo[]");


  // different types
  $a2 = [new Foo(), new Foo()];
  $b2 = [new Foo(), ...returnBooArray()];
  exprtype($b2, "mixed[]");


  // with two unpack
  $a3 = [new Foo(), new Foo()];
  $b3 = [new Foo(), new Foo()];
  $c3 = [...$a3, ...$b3];
  exprtype($c3, "\Foo[]");


  // with two unpack with different type
  $a4 = [new Foo(), new Foo()];
  $c4 = [...$a4, ...returnBooArray()];
  exprtype($c4, "mixed[]");


  // one unpack
  $a5 = [new Foo(), new Foo()];
  $b5 = [...$a5];
  exprtype($b5, "\Foo[]");


  // with two unpack and just type
  $a6 = [new Foo(), new Foo()];
  $b6 = [new Foo(), new Foo()];
  $c6 = [...$a6, new Foo(),...$b6];
  exprtype($c6, "\Foo[]");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestPropertyTypeHints(t *testing.T) {
	code := `<?php
class Too {}
class Boo {}

// simple use
class Foo {
  public int $int;
  public array $array;
  public Boo $boo;

  public static float $float;
  public static object $object;
  public static Too $too;
}

$f = new Foo();

exprtype($f->int, "int");
exprtype($f->array, "mixed[]");
exprtype($f->boo, "\Boo");

exprtype(Foo::$float, "float");
exprtype(Foo::$object, "object");
exprtype(Foo::$too, "\Too");


// with PHPDoc
class Poo {
  /** @var float $int */
  public static int $int;

  /** @var array $callable */
  public object $object;

  /** @var array $array */
  public array $array;

  /** @var float|int $a */
  public static int $a;

  /** @var int $b */
  public int $b;

  /** @var Foo|Too $c */
  public Boo $c;

  /** @var Boo $d */
  public Boo $d;
}

$p = new Poo();

exprtype(Poo::$int, "float|int");
exprtype($p->object, "mixed[]|object");
exprtype($p->array, "mixed[]");
exprtype(Poo::$a, "float|int");
exprtype($p->b, "int");
exprtype($p->c, "\Boo|\Foo|\Too");
exprtype($p->d, "\Boo");


// with default value
class Roo {
  /** @var float $a */
  public static int $a = 10;

  /** @var array $b */
  public array $b = [1,2,3];

  public bool $c = true;

  public static string $d = "Hello";
}

$r = new Roo();

exprtype(Roo::$a, "float|int");
exprtype($r->b, "int[]|mixed[]");
exprtype($r->c, "bool");
exprtype(Roo::$d, "string");
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestArrowFunction(t *testing.T) {
	code := `<?php
function f() {
   $value = 10;
   $_ = fn($x) => $value = "probably now $value has type int|string";
   // but, no
   exprtype($value, "precise int"); // Ok, see specification
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestClosureCallbackArgumentsTypes(t *testing.T) {
	code := `<?php
function usort($array, $callback) {}
function uasort($array, $callback) {}
function array_map($callback, $array, $_ = null) {}
function array_walk($callback, $array) {}
function array_walk_recursive($callback, $array) {}
function array_filter($array, $callback) {}
function array_reduce($array, $callback) {}
function some_function_without_model($callback, $array) {}

class Foo { public function f() {} }
class Boo { public function b() {} }

/** @return Foo[] */
function return_foo() { return []; }

/** @return Boo[] */
function return_boo() { return []; }

$foo_array = return_foo();
$boo_array = return_boo();

usort($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
});

uasort($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
});

array_map(function($a) {
  exprtype($a, "\Foo");
}, $foo_array);

array_walk($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
});

// wrong number of args
array_walk($foo_array, function($a, $b, $c) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
  exprtype($c, "mixed");
});

array_walk_recursive($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
});

array_filter($foo_array, function($a) {
  exprtype($a, "\Foo");
});

array_reduce($foo_array, function($carry, $item) {
  exprtype($carry, "\Foo");
  exprtype($item, "\Foo");
});

// mixed array, but function args with type hints
$mixed_arr = [];
usort($mixed_arr, function(Foo $a, Foo $b) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
});

$mixed_arr_2 = [];
usort($mixed_arr_2, function(int $a, int $b) {
  exprtype($a, "int");
  exprtype($b, "int");
});

// mixed array, but not all function args have type hints
$mixed_arr_3 = [];
usort($mixed_arr_3, function(Foo $a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
});

// non mixed array, but function args with type hints
$non_mixed_arr = [1, 2, 3];
usort($non_mixed_arr, function(int $a, int $b) {
  exprtype($a, "int");
  exprtype($b, "int");
});

// non mixed array, but not all function args have type hints
$non_mixed_arr_2 = [new Foo, new Foo, new Foo];
usort($non_mixed_arr_2, function(Foo $a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
});

some_function_without_model(function($b) {
  exprtype($b, "mixed");
}, $d);

// Not supported
function callback($a) {
  exprtype($a, "mixed");
}

// Not supported
$callback = function($a) {
  exprtype($a, "mixed");
};

array_map($callback, $d);
array_map('callback', $d);
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestClosureCallbackArgumentsPossibleErrorVariations(t *testing.T) {
	code := `<?php
function usort($array, $callback) {}
function uasort($array, $callback) {}
function array_map($callback, $array, $_ = null) {}
function array_walk($callback, $array) {}
function array_walk_recursive($callback, $array) {}
function array_filter($array, $callback) {}
function array_reduce($array, $callback) {}

class Foo { public function f() {} }

/** @return Foo[] */
function return_foo() { return []; }

$foo_array = return_foo();

// more arguments than necessary
usort($foo_array, function($a, $b, $c) {
  exprtype($a, "\Foo");
  exprtype($b, "\Foo");
  exprtype($c, "mixed");
});

// less arguments than necessary
uasort($foo_array, function($a) {
  exprtype($a, "\Foo");
});

// more arguments than necessary
array_map(function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
}, $foo_array);

// less arguments than necessary
array_map(function() {

}, $foo_array);

// more arguments than necessary
array_walk($foo_array, function($a, $b, $c) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
  exprtype($c, "mixed");
});

// less arguments than necessary
array_walk($foo_array, function() {
});

// more arguments than necessary
array_filter($foo_array, function($a, $b) {
  exprtype($a, "\Foo");
  exprtype($b, "mixed");
});

// less arguments than necessary
array_filter($foo_array, function() {
});

// more arguments than necessary
array_reduce($foo_array, function($carry, $item, $c) {
  exprtype($carry, "\Foo");
  exprtype($item, "\Foo");
  exprtype($c, "mixed");
});

// less arguments than necessary
array_reduce($foo_array, function() {
});
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestNestedClosureCallback(t *testing.T) {
	code := `<?php
function array_map($callback, array $arr1, array $_ = null) {}
function usort($array, $callback) {}

class Foo { public function f() {} }
/** @return Foo[][] */

function return_foo() { return []; }

$foo_array = return_foo();

array_map(function($a) {
  exprtype($a, "\Foo[]");
  array_map(function($a) {
    exprtype($a, "\Foo");
  }, $a);
  exprtype($a, "\Foo[]");
}, $foo_array);

usort($foo_array, function($a, $b) {
  exprtype($a, "\Foo[]");
  exprtype($b, "\Foo[]");

  array_map(function($a) {
    exprtype($a, "\Foo");
  }, $a);

  array_map(function($b) {
    exprtype($b, "\Foo");
  }, $b);

  exprtype($a, "\Foo[]");
  exprtype($b, "\Foo[]");
});
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestMemberTypeInPHPDoc(t *testing.T) {
	code := `<?php
class Foo {
	const BAR = 5;
	const BAZ = 10;
}

/**
 * @return \Foo::BAR|\Foo::BAZ
 */
function f() {}

function f2() {
	exprtype(f(), "mixed");
}

`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestTypeWithAssignOperators(t *testing.T) {
	code := `<?php
function g($x) {
	exprtype($x <<= 5, 'precise int');
	exprtype($x, 'precise int');

	exprtype($x .= 'abc', 'precise string');
	exprtype($x, 'precise string');

	exprtype($x >>= 5, 'precise int');
	exprtype($x, 'precise int');
}

function f() {
	$a = 10;
	$a += 12.5;
	exprtype($a, "float");
	$a = 10;
	$a += 12;
	exprtype($a, "int");

	$a = 10;
	$a -= 12.5;
	exprtype($a, "float");
	$a = 10;
	$a -= 12;
	exprtype($a, "int");

	$a = "Hello";
	$a .= 12.5;
	exprtype($a, "precise string");
	$a = "Hello";
	$a .= " World";
	exprtype($a, "precise string");

	$a = 5;
	$a /= 5.5;
	exprtype($a, "float");
	$a = 5;
	$a /= 5;
	exprtype($a, "int");

	$a = 5;
	$a *= 5.5;
	exprtype($a, "float");
	$a = 5;
	$a *= 5;
	exprtype($a, "int");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestMagicConstants(t *testing.T) {
	code := `<?php
class Foo {}

function f() {
	$line = __LINE__;
	exprtype($line, "precise int");

	$file = __FILE__;
	exprtype($file, "precise string");

	$dir = __DIR__;
	exprtype($dir, "precise string");

	$function = __FUNCTION__;
	exprtype($function, "precise string");

	$class = __CLASS__;
	exprtype($class, "precise string");

	$trait = __TRAIT__;
	exprtype($trait, "precise string");

	$method = __METHOD__;
	exprtype($method, "precise string");

	$namespace = __NAMESPACE__;
	exprtype($namespace, "precise string");

	$className = Foo::class;
	exprtype($className, "precise string");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestNullCoalesceType(t *testing.T) {
	code := `<?php
class Foo {}

function f() {
	$a = 10;
	$b = "Hello";

	$c = $a ?? $b;
	exprtype($c, "precise int|string");

	$f = new Foo();

	$s = $c ?? $f;
	exprtype($s, "\Foo|int|string");

	$e = 10.5;

	$e ??= $s;
	exprtype($e, "\Foo|float|int|string");
  }
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestNewFunctionReturnExprType(t *testing.T) {
	code := `<?php
class Foo {}

/** @return int */
function f1() {
	return 5;
}
exprtype(f1(), "int");


/** @return int */
function f2(): float {
	return 5;
}
exprtype(f2(), "float|int");


/** Without @return */
function f3(): float {
	if (1) {
		return new Foo;
	}
	return 5;
}
exprtype(f3(), "float");


/** Without @return and type hint */
function f4() {
	if (1) {
		return 10;
	}
	return new Foo;
}
exprtype(f4(), "\Foo|int");


/** Without @return and type hint and return */
function f5() {
	
}
exprtype(f5(), "void");


/** @return int[] */
function f6(): array {
	return [1,2,4];
}
exprtype(f6(), "int[]");


/** @return int[] */
function f7(): array {
	return f6();
}
exprtype(f7(), "int[]");


/** @return Foo */
function f8(): object {
	return new Foo;
}
exprtype(f8(), "\Foo");
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestClosureExprType(t *testing.T) {
	code := `<?php
class Foo {
  public function method() {}
}

function func() {
  $f = function(): Foo { return new Foo; };
  $foo = $f();
  
  exprtype($f, "\Closure$(exprtype.php,func):7$");
  exprtype($foo, "\Foo");
}

function func1() {
  $f1 = function(): float {
    if (1) {
	  return new Foo;
	}
    return 10; 
  };

  $foo1 = $f1();

  exprtype($f1, "\Closure$(exprtype.php,func1):15$");
  exprtype($foo1, "float");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestNullableInTupleExprType(t *testing.T) {
	code := `<?php
class Boo {}

class Foo {
	/**
	 * @return tuple(?Foo, int)
	 */
	public static function staticMethodNullable() {}

	/**
	 * @return tuple(?Boo, int)
	 */
	public static function staticMethodNullableBoo() {}

	/**
	 * @return tuple(?self, int)
	 */
	public static function staticMethodSelfNullable() {}
}

function f1() {
	list($a, $_) = Foo::staticMethodNullable();
	list($b, $_) = Foo::staticMethodNullableBoo();
	list($c, $_) = Foo::staticMethodSelfNullable();
	exprtype($a, "\Foo|null");
	exprtype($b, "\Boo|null");
	exprtype($c, "\Foo|null");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestSelfStaticInTupleExprType(t *testing.T) {
	code := `<?php
class Foo {
	/**
	 * @return tuple(static, int)
	 */
	public static function staticMethodStatic() {}
	
	/**
	 * @return tuple($this, int)
	 */
	public static function staticMethodThis() {}
	
	/**
	 * @return tuple(self, int)
	 */
	public static function staticMethodSelf() {}
}

function f1() {
    list($a, $_) = Foo::staticMethodStatic();
    list($b, $_) = Foo::staticMethodThis();
    list($c, $_) = Foo::staticMethodSelf();
    exprtype($a, "\Foo");
    exprtype($b, "\Foo");
    exprtype($c, "\Foo");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestClassesInTupleExprType(t *testing.T) {
	code := `<?php
namespace Boo {
	class B {}
	class C {}
}

namespace Foo {
	use Boo\B;
	use Boo\C as ClassFromBoo;

	class F extends ClassFromBoo {
		/**
		 * @return tuple(parent, int)
		 */
		public static function method() {}
	}
	
	/**
	 * @return tuple(?B, integer)
	 */
	function f() {}

	/**
	 * @return tuple(ClassFromBoo, int)
	 */
	function f1() {}
	
	function f2() {
		list($a, $_) = f();
		list($b, $_) = f1();
		list($c, $_) = F::method();
		
		exprtype($a, "\Boo\B|null");
		exprtype($b, "\Boo\C");
		exprtype($c, "\Boo\C");
	}
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestTupleWithArray(t *testing.T) {
	code := `<?php
/**
 * @return tuple(array, array)
 */
function f() {
    return [[],[]];
}

exprtype(f()[0], "mixed[]");
exprtype(f()[1], "mixed[]");
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestMultiDimensionalArray(t *testing.T) {
	code := `<?php
class Foo {
	public function f(): self {}
}

function f() {
	$a = [];    
	
	$a[] = [1,2,3];
	$a[] = ["1","2","3"];
	exprtype($a, "int[][]|string[][]");

	foreach ($a as $val) {
		exprtype($val, "int[]|string[]");
		foreach ($val as $val2) {
			exprtype($val2, "int|string");
		}
	}

	$a[1][2] = new Foo;

	foreach ($a as $val) {
		exprtype($val, "\Foo[]|int[]|string[]");
		foreach ($val as $val2) {
			exprtype($val2, "\Foo|int|string");
			$a = $val2->f();
			exprtype($a, "\Foo");
		}
	}
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestCallableDoc(t *testing.T) {
	code := `<?php
class Foo {
  /**
   * @return int
   */
  public function method(): int { return 0; }
}

class Boo {
  /**
   * @return int
   */
  public function method(): int { return 0; }
}

/**
 * @param callable(): Foo $s
 */
function f1(callable $s) {
  $a = $s();
  exprtype($a, "\Foo");
}

/**
 * @param callable(Foo): Foo $s
 */
function f2(callable $s) {
  $a = $s(new Foo);
  exprtype($a, "\Foo");
}

/**
 * @param callable(int): Foo $s
 */
function f3(callable $s) {
  $a = $s(10);
  exprtype($a, "\Foo");
}

/**
 * @param callable(int, string): Foo|Boo $s
 */
function f4(callable $s) {
  $a = $s(10, "ss");
  exprtype($a, "\Boo|\Foo");
}

/**
 * @param callable(): callable(): Foo $s
 * @param callable(): callable(): Foo|Boo $s1
 */
function f5(callable $s, callable $s1) {
  $a = $s();
  $a1 = $s1();
  exprtype($a, "\Closure$():Foo");
  exprtype($a1, "\Closure$():Foo/Boo");
  $b = $a();
  $b1 = $a1();
  exprtype($b, "\Foo");
  exprtype($b1, "\Boo|\Foo");
}

/**
 * @return callable(): callable(): Foo
 */
function f6(): callable {
  return function() { return function() { return new Foo; }; };
}

function f7() {
  $a = f6();
  exprtype($a, "\Closure$():callable(): Foo|callable");
  $b = $a();
  exprtype($b, "\Closure$():Foo");
  $c = $b();
  exprtype($c, "\Foo");
}

function f8() {
  /**
   * @var callable(): Foo $a
   */
  $a = null;

  $b = $a();
  exprtype($b, "\Foo");
}

/**
* @var callable(): Foo $a
*/
$a = null;
$b = $a();
exprtype($b, "\Foo");

/**
 * @param callable(int, string) $s
 */
function f9(callable $s) {
  $a = $s(10, "ss");
  exprtype($a, "mixed");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestDifferentArraySyntax(t *testing.T) {
	code := `<?php
/**
 * @param array<Foo> $arr
 * @param list<Foo> $arr1
 * @param non-empty-array<Foo> $arr2
 * @param non-empty-list<Foo> $arr3
 * @param unknown-type-list<Foo> $arr4
 * @param iterable<Foo> $arr5
 */
function f($arr, $arr1, $arr2, $arr3, $arr4, $arr5) {
  exprtype($arr, "\Foo[]");
  exprtype($arr1, "\Foo[]");
  exprtype($arr2, "\Foo[]");
  exprtype($arr3, "\Foo[]");
  exprtype($arr4, "\Foo[]");
  exprtype($arr5, "\Foo[]");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func TestArrayTypeCast(t *testing.T) {
	code := `<?php
class Foo {}

/**
 * @return Foo[]
 */
function f() {
  return [];
}

function f1() {
  $a = (array) f();
  exprtype($a, "\Foo[]|mixed[]");
  exprtype($a[0], "\Foo|mixed");

  $b = (array) 10;
  exprtype($b, "int|mixed[]");

  $c = (array) [1, "s"];
  exprtype($c, "mixed[]");
}
`
	runExprTypeTest(t, &exprTypeTestParams{code: code})
}

func runExprTypeTest(t *testing.T, params *exprTypeTestParams) {
	exprTypeTestImpl(t, params, false)
}

func exprTypeTestImpl(t *testing.T, params *exprTypeTestParams, kphp bool) {
	config := linter.NewConfig()
	config.Checkers.AddBlockChecker(func(ctx *linter.BlockContext) linter.BlockChecker {
		return &exprTypeCollector{ctx: ctx}
	})
	config.KPHP = kphp
	l := linter.NewLinter(config)

	if params.stubs != "" {
		l.InitStubs(func(ch chan workspace.FileInfo) {
			ch <- workspace.FileInfo{
				Name:     "stubs.php",
				Contents: []byte(params.stubs),
			}
		})
	}
	linttest.ParseTestFile(t, l, "exprtype.php", params.code)

	l.MetaInfo().SetIndexingComplete(true)

	// Reset results map and run expr type collector.
	exprTypeResult = map[ir.Node]types.Map{}
	result := linttest.ParseTestFile(t, l, "exprtype.php", params.code)

	// Check that collected types are identical to the expected types.
	// We need the second walker to pass *testing.T parameter to
	// the walker that does the comparison.
	walker := exprTypeWalker{t: t}
	result.RootNode.Walk(&walker)
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

func (w *exprTypeWalker) LeaveNode(n ir.Node) {}

func (w *exprTypeWalker) EnterNode(n ir.Node) bool {
	call, ok := n.(*ir.FunctionCallExpr)
	if ok && utils.NameNodeEquals(call.Function, `exprtype`) {
		checkedExpr := call.Arg(0).Expr
		expectedType := call.Arg(1).Expr.(*ir.String).Value
		actualType, ok := exprTypeResult[checkedExpr]
		if !ok {
			w.t.Fatalf("no type found for %s expression", irutil.FmtNode(checkedExpr))
		}
		want := makeType(expectedType)
		have := testTypesMap{
			Types:   actualType.String(),
			Precise: actualType.IsPrecise(),
		}
		if diff := cmp.Diff(have, want); diff != "" {
			line := ir.GetPosition(checkedExpr).StartLine
			w.t.Errorf("line %d: type mismatch for %s (-have +want):\n%s",
				line, irutil.FmtNode(checkedExpr), diff)
		}
		return false
	}

	return true
}

type exprTypeCollector struct {
	ctx *linter.BlockContext
	linter.BlockCheckerDefaults
}

func (c *exprTypeCollector) AfterEnterNode(n ir.Node) {
	if !c.ctx.ClassParseState().Info.IsIndexingComplete() {
		return
	}

	call, ok := n.(*ir.FunctionCallExpr)
	if !ok || !utils.NameNodeEquals(call.Function, `exprtype`) {
		return
	}
	checkedExpr := call.Arg(0).Expr

	// We need to clone a types map because if it belongs to a var
	// or some other symbol those type can be volatile we'll get
	// unexpected results.
	typ := c.ctx.ExprType(checkedExpr).Clone()

	exprTypeResultMu.Lock()
	exprTypeResult[checkedExpr] = typ
	exprTypeResultMu.Unlock()
}
