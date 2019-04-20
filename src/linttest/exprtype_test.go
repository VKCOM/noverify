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
		{`[1, 2]`, "int[]"},
		{`[1.4, 3.5]`, "float[]"},
		{`["1", "5"]`, "string[]"},

		{`[$int, $int]`, "array"}, // TODO: could be int[]

		{`[11][0]`, "int"},
		{`["11"][0]`, "string"},
		{`[1.4][0]`, "float"},
	}

	local := `$int = 10`
	runExprTypeTest(t, &exprTypeTestContext{local: local}, tests)
}

func TestExprTypeMulti(t *testing.T) {
	tests := []exprTypeTest{
		{`$cond ? 1 : 2`, "int"},
		{`$int_or_float`, "int|float"},
		{`$int_or_float`, "float|int"},
		{`$cond ? 10 : "123"`, "int|string"},
		{`$cond ? ($int_or_float ? 10 : 10.4) : (bool)1`, "int|float|bool"},
	}

	global := `<?php
$cond = "true";
$int_or_float = 10;
if ($cond) {
  $int_or_float = 10.5;
}
`
	runExprTypeTest(t, &exprTypeTestContext{global: global}, tests)
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

func runExprTypeTest(t *testing.T, ctx *exprTypeTestContext, tests []exprTypeTest) {
	if ctx == nil {
		ctx = &exprTypeTestContext{}
	}

	meta.ResetInfo()
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
