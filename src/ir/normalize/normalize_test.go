package normalize

import (
	"errors"
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/php7"
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
			`Foo::$x; Foo::myconst;`,
		},
	}

	conf := Config{CurrentClass: `Foo`}
	for _, test := range tests {

		list, err := parseStmtList(test.orig)
		if err != nil {
			t.Errorf("parse `%s`: %v", test.orig, err)
			continue
		}
		normalized := FuncBody(conf, nil, list)
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

func runNormalizeTests(t *testing.T, tests []normalizationTest) {
	t.Helper()

	conf := Config{CurrentClass: `Foo`}
	for _, test := range tests {
		n, _, err := parseutil.ParseStmt([]byte(test.orig))
		if err != nil {
			t.Errorf("parse `%s`: %v", test.orig, err)
			return
		}
		irNode := irconv.ConvertNode(n)
		normalized := FuncBody(conf, nil, []ir.Node{irNode})
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
		irNode2 := irconv.ConvertNode(n2)
		normalized2 := FuncBody(conf, nil, []ir.Node{irNode2})
		have2 := strings.TrimSuffix(irfmt.Node(normalized2[0]), ";")
		if have != have2 {
			t.Errorf("second normalization of `%s`:\nhave: %s\nwant: %s",
				test.orig, have2, have)
		}
	}
}

func TestNormalizeExpr(t *testing.T) {
	runNormalizeTests(t, []normalizationTest{
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
		{`list($x, $y) = f()`, `[$v0, $v1] = f()`},

		{`$x ? $x : $y`, `$v0 ?: $v1`},
		{`f() ? f() : $y`, `f() ? f() : $v0`},
	})
}

func TestNormalizeExprAfterIndexing(t *testing.T) {
	linttest.ParseTestFile(t, "defs.php", `<?php
const ZERO = 0;
const HELLO_WORLD = 'hello, world';
const LOCALHOST = "127.0.0.1";
`)
	meta.SetIndexingComplete(true)
	defer meta.SetIndexingComplete(false)

	runNormalizeTests(t, []normalizationTest{
		{`ZERO`, `0`},
		{`HELLO_WORLD`, `'hello, world'`},
		{`LOCALHOST`, `'127.0.0.1'`},
	})
}

func parseStmtList(s string) ([]ir.Node, error) {
	source := "<?php " + s
	p := php7.NewParser([]byte(source))
	p.Parse()
	if len(p.GetErrors()) != 0 {
		return nil, errors.New(p.GetErrors()[0].String())
	}
	rootIR := irconv.ConvertRoot(p.GetRootNode())
	return rootIR.Stmts, nil
}
