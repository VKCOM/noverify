package linttest_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
)

// Tests in this file make it less likely that type solving will break
// without being noticed.

// TODO(quasilyte): better handling of an `empty_array` type.
// Now it's resolved to `array` for expressions that have multiple empty_array.

func TestExprTypeMalformedPhpdoc(t *testing.T) {
	tests := []exprTypeTest{
		{`return_mixed(0)`, ``},
		{`return_int(0)`, `int`},
	}

	global := `<?php
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

	global := `
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

		{`[]`, "array"},
		{`[1, "a", 4.5]`, "array"},

		{`$int`, "int"},
		{`$float`, "float"},
		{`$string`, "string"},
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
		{`[]`, `array`}, // Should never be "empty_array" after resolving
		{`[[]]`, `array`},

		{`[1, 2]`, "int[]"},
		{`[1.4, 3.5]`, "float[]"},
		{`["1", "5"]`, "string[]"},

		{`[$int, $int]`, "array"}, // TODO: could be int[]

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
	}

	global := `<?php
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

	global := `
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

		// TODO:
		//
		// {`Gopher::$name`, "string"},
		// {`Gopher::POWER`, "int"},
		// {`$magic->int`, ""},
	}

	global := `
class Gopher {
  /** @var string */
  public static $name = "unnamed";

  constant POWER = 9001; // It's over 9000
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
		{`get_array()`, `array`},
		{`get_array_or_null()`, `array|null`},
		{`get_null_or_array()`, `array|null`},
	}

	global := `<?php
function define($name, $value) {}
define('null', 0);

class Foo {}

function get_ints() {
	$a = []; // "empty_array"
	$a[0] = 1;
	$a[1] = 2;
	return $a; // Should be resolved to just int[]
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
	local := `$test = new \NS\Test();`
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
	}

	stubs := `<?php
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
		root, _ := linttest.ParseTestFile(t, "exprtype_global.php", "<?php\n"+ctx.global)
		root.Walk(&gw)
	}
	sources := exprTypeSources(ctx, tests, gw.globals)
	linttest.ParseTestFile(t, "exprtype.php", sources)
	meta.SetIndexingComplete(true)

	for i, test := range tests {
		fn, ok := meta.Info.GetFunction(fmt.Sprintf("\\f%d", i))
		if !ok {
			t.Errorf("missing f%d info", i)
			continue
		}
		have := solver.ResolveTypes(fn.Typ, make(map[string]struct{}))
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

func (gw *globalsWalker) GetChildrenVisitor(string) walker.Visitor { return gw }
func (gw *globalsWalker) LeaveNode(walker.Walkable)                {}
