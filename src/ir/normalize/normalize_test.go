package normalize

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/z7zmey/php-parser/pkg/cfg"
	phperrors "github.com/z7zmey/php-parser/pkg/errors"
	"github.com/z7zmey/php-parser/pkg/parser"
	"github.com/z7zmey/php-parser/pkg/version"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parseutil"
)

func TestNormalizeStmtList(t *testing.T) {
	tests := []struct {
		orig string
		want string
	}{
		// Global vars should not be renamed.
		{
			`global $x; echo $x;`,
			`global $x; echo $x;`,
		},

		// Swap if then/else for negated conditions.
		{
			`if (!$x) { a(); } else { b(); }`,
			`if ($v0) { b(); } else { a(); }`,
		},

		// $x++ transformed into ++$x in stmt context.
		{
			`$x++; $x--;`,
			`++$v0; --$v0;`,
		},
		{
			`for (; ; $x++);`,
			`for (; ; ++$v0);`,
		},
		// ...but not in expression context.
		{
			`$x = $y++;`,
			`$v0 = $v1++;`,
		},

		// array_push() of 2 arguments is translated into []= assignment.
		{
			`array_push($a, 10);`,
			`$v0[] = 10;`,
		},

		// Some statements can be re-ordered (sorted).
		{
			`$x = 0; $y = 'abc';`,
			`$v0 = 'abc'; $v1 = 0;`,
		},
		{
			`require 'b'; require 'a';`,
			`require 'a'; require 'b';`,
		},

		{
			`f(array(), "1$x$y/23");`,
			`f([], '1' . $v0 . $v1 . '/23');`,
		},

		{
			`self::$x; self::myconst;`,
			`\Foo::$x; \Foo::myconst;`,
		},
	}

	conf := Config{}
	l := linter.NewLinter(linter.NewConfig())
	st := &meta.ClassParseState{Info: l.MetaInfo(), CurrentClass: `\Foo`}
	for _, test := range tests {

		list, err := parseStmtList(test.orig)
		if err != nil {
			t.Errorf("parse `%s`: %v", test.orig, err)
			continue
		}
		normalized := FuncBody(st, conf, nil, list)
		have := strings.ReplaceAll(irfmt.Node(&ir.StmtList{Stmts: normalized}), "\n", "")
		have = strings.ReplaceAll(have, "  ", " ")
		want := `{ ` + test.want + `}`
		if have != want {
			t.Errorf("normalize `%s`:\nhave: %s\nwant: %s",
				test.orig, have, want)
			continue
		}
	}
}

type normalizationTest struct {
	orig string
	want string
}

func runNormalizeTests(t *testing.T, l *linter.Linter, tests []normalizationTest) {
	t.Helper()

	conf := Config{}
	st := &meta.ClassParseState{Info: l.MetaInfo(), CurrentClass: `\Foo`}
	irConverter := irconv.NewConverter(nil)
	for _, test := range tests {
		n, _, err := parseutil.ParseStmt([]byte(test.orig))
		if err != nil {
			t.Errorf("parse `%s`: %v", test.orig, err)
			return
		}
		irNode := irConverter.ConvertNode(n)
		normalized := FuncBody(st, conf, nil, []ir.Node{irNode})
		have := strings.TrimSuffix(irfmt.Node(normalized[0]), ";")
		if have != test.want {
			t.Errorf("normalize `%s`:\nhave: %s\nwant: %s",
				test.orig, have, test.want)
			return
		}
		n2, _, err := parseutil.ParseStmt([]byte(have))
		if err != nil {
			t.Errorf("parse normalized `%s`: %v", test.orig, err)
			return
		}
		irNode2 := irConverter.ConvertNode(n2)
		normalized2 := FuncBody(st, conf, nil, []ir.Node{irNode2})
		have2 := strings.TrimSuffix(irfmt.Node(normalized2[0]), ";")
		if have != have2 {
			t.Errorf("second normalization of `%s`:\nhave: %s\nwant: %s",
				test.orig, have2, have)
		}
	}
}

func TestNormalizeExpr(t *testing.T) {
	l := linter.NewLinter(linter.NewConfig())
	runNormalizeTests(t, l, []normalizationTest{
		{`new T`, `new T()`},

		{`"$x"`, `'' . $v0`},

		{`$x`, `$v0`},
		{`echo $x`, `echo $v0`},

		{`!!$x`, `(bool)$v0`},

		// Commutative operands are sorted if they don't have side effects.
		{`0 == $x`, `$v0 == 0`},
		{`$x === "abc"`, `"abc" === $v0`},
		{`$x + $y + $x`, `$v0 + $v0 + $v1`},

		{`"abc"`, `'abc'`},

		{`NULL`, `null`},
		{`MyConst`, `MyConst`},

		{`is_null($x)`, `$v0 === null`},

		{`is_integer($v)`, `is_int($v0)`},
		{`is_real($v)`, `is_float($v0)`},

		{`$x = $x + 2`, `$v0 += 2`},
		{`$x = $x + 1`, `++$v0`},
		{`$x = $x - 1`, `--$v0`},

		{`array(1, 2)`, `[1, 2]`},
		{`list($x, $y) = f()`, `[0 => $v0, 1 => $v1] = f()`},

		{`$x ? $x : $y`, `$v0 ?: $v1`},
		{`f() ? f() : $y`, `f() ? f() : $v0`},

		// Const-folded.
		{`1 + 5`, `6`},
		{`(1 + 5) + 1`, `7`},
		{`1 + 2 * 4`, `9`},
		{`(1 + 2) * 4`, `12`},
		{`5 - 6`, `-1`},
		{`(5 - 6) * 8 + 6`, `-2`},
		{`5.5 + 3`, `8.5`},
		{`3 + 5.5`, `8.5`},
		{`5.5 + 3.5`, `9`},
		{`5.5 - 3`, `2.5`},
		{`8 - 5.5`, `2.5`},
		{`5.5 - 3.5`, `2`},
		{`5.5 * 3`, `16.5`},
		{`3 * 5.5`, `16.5`},
		{`1.5 * 1.5`, `2.25`},
		{`!true`, `false`},
		{`!false`, `true`},
		{`!!false`, `false`},
		{`true || false`, `true`},
		{`true && false`, `false`},
		{`true && true`, `true`},
		{`true || true`, `true`},
		{`!true && (!!false || true)`, `false`},
		{`true and false`, `false`},
		{`true or false`, `true`},
		{`0b1000 | 0b1`, `9`},
		{`0b10 & 0b11`, `2`},
		{`"Hello " . "World!"`, `'Hello World!'`},
		{`"Hello " . "World" . "!"`, `'Hello World!'`},

		// List assignments.
		{`list(, $x) = f()`, `[1 => $v0] = f()`},
		{`list(, $x, , $y) = f()`, `[1 => $v0, 3 => $v1] = f()`},
		{`[5 => $x] = f()`, `[5 => $v0] = f()`},
		{`[$x, ] = f()`, `[0 => $v0] = f()`},
	})
}

func TestNormalizeExprAfterIndexing(t *testing.T) {
	l := linter.NewLinter(linter.NewConfig())
	linttest.ParseTestFile(t, l, "defs.php", `<?php
const ZERO = 0;
const HELLO_WORLD = 'hello, world';
const LOCALHOST = "127.0.0.1";

class Foo {
  const FOO_VALUE = 53.001122334455665;
}
`)
	l.MetaInfo().SetIndexingComplete(true)

	runNormalizeTests(t, l, []normalizationTest{
		{`ZERO`, `0`},
		{`HELLO_WORLD`, `'hello, world'`},
		{`LOCALHOST`, `'127.0.0.1'`},
		{`self::FOO_VALUE`, `53.001122334455665`},
	})
}

func TestMagicConstFold(t *testing.T) {
	l := linter.NewLinter(linter.NewConfig())
	linttest.ParseTestFile(t, l, "files/file.php", `<?php
namespace Boo;

const NAMESPACE_NAME = __NAMESPACE__;
const FILENAME = __FILE__;
const DIR = __DIR__;
const DIR_WITH_FUNC = dirname(__FILE__);

const CUSTOM_DIR = __DIR__ . "/file2.php";
const CUSTOM_DIR_2 = dirname(__FILE__) . "/file2.php";

class Foo {
  const CLASSNAME = __CLASS__;
  const LINE = __LINE__;
}
`)
	l.MetaInfo().SetIndexingComplete(true)

	runNormalizeTests(t, l, []normalizationTest{
		{`\Boo\NAMESPACE_NAME`, `'\Boo'`},
		{`\Boo\FILENAME`, `'files/file.php'`},
		{`\Boo\DIR`, `'files'`},
		{`\Boo\DIR_WITH_FUNC`, `'files'`},
		{`\Boo\CUSTOM_DIR`, `'files/file2.php'`},
		{`\Boo\CUSTOM_DIR_2`, `'files/file2.php'`},
		{`\Boo\Foo::CLASSNAME`, `'\Boo\Foo'`},
		{`\Boo\Foo::LINE`, `14`},
	})
}

func parseStmtList(s string) ([]ir.Node, error) {
	source := "<?php " + s
	// p := php7.NewParser([]byte(source))
	// p.Parse()
	// if len(p.GetErrors()) != 0 {
	// 	return nil, errors.New(p.GetErrors()[0].String())
	// }

	phpVersion, err := version.New("7.4")
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}

	var parserErrors []*phperrors.Error
	rootNode, err := parser.Parse([]byte(source), cfg.Config{
		Version: phpVersion,
		ErrorHandlerFunc: func(e *phperrors.Error) {
			parserErrors = append(parserErrors, e)
		},
	})
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return nil, errors.New(parserErrors[0].String())
	}

	rootIR := irconv.ConvertNode(rootNode).(*ir.Root)

	return rootIR.Stmts, nil
}
