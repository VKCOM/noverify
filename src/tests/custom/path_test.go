package custom

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

func init() {
	linter.RegisterBlockChecker(func(ctx *linter.BlockContext) linter.BlockChecker {
		return &pathTester{ctx: ctx}
	})
}

func TestPathIfElse(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
if ($cond1) {
  echo $_if;
} elseif ($cond2) {
  echo $_elseif1;
} else if ($cond3) {
  echi $_elseif2;
} else {
  echo $_else;
}
`)
	test.Expect = []string{
		`$_if (cond=true) : *node.Root/*stmt.If/*stmt.StmtList/*stmt.Echo/*node.SimpleVar`,
		`$_elseif1 (cond=true) : *node.Root/*stmt.If/*stmt.StmtList/*stmt.Echo/*node.SimpleVar`,
		`$_elseif2 (cond=true) : *node.Root/*stmt.If/*stmt.StmtList/*stmt.Expression/*node.SimpleVar`,
		`$_else (cond=true) : *node.Root/*stmt.If/*stmt.Else/*stmt.StmtList/*stmt.Echo/*node.SimpleVar`,
	}
	linttest.RunFilterMatch(test, "pathTest")
}

func TestPathArgument(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
f($_arg1, [$_arg2], $a + $_arg3);
`)
	test.Expect = []string{
		`$_arg1 (cond=false) : *node.Root/*stmt.Expression/*expr.FunctionCall/*node.Argument/*node.SimpleVar`,
		`$_arg2 (cond=false) : *node.Root/*stmt.Expression/*expr.FunctionCall/*node.Argument/*expr.Array/*node.SimpleVar`,
		`$_arg3 (cond=false) : *node.Root/*stmt.Expression/*expr.FunctionCall/*node.Argument/*binary.Plus/*node.SimpleVar`,
	}
	linttest.RunFilterMatch(test, "pathTest")
}

func TestPathTernary(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
echo $_ternary_cond ? $_ternary_true : $_ternary_false;
`)
	test.Expect = []string{
		`$_ternary_cond (cond=true) : *node.Root/*stmt.Echo/*expr.Ternary/*node.SimpleVar`,
		`$_ternary_true (cond=true) : *node.Root/*stmt.Echo/*expr.Ternary/*node.SimpleVar`,
		`$_ternary_false (cond=true) : *node.Root/*stmt.Echo/*expr.Ternary/*node.SimpleVar`,
	}
	linttest.RunFilterMatch(test, "pathTest")
}

func TestPathTopLevel(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
if ($_cond) {
  echo $_x1;
  switch (true) {
  case $_case:
    echo $_x2;
    break;
  }
}

do {
  $_at_least_once;
} while ($_do_while_cond);
`)
	test.Expect = []string{
		`$_cond (cond=true) : *node.Root/*stmt.If/*node.SimpleVar`,
		`$_x1 (cond=true) : *node.Root/*stmt.If/*stmt.StmtList/*stmt.Echo/*node.SimpleVar`,
		`$_case (cond=true) : *node.Root/*stmt.If/*stmt.StmtList/*stmt.Switch/*node.SimpleVar`,
		`$_x2 (cond=true) : *node.Root/*stmt.If/*stmt.StmtList/*stmt.Switch/*stmt.Echo/*node.SimpleVar`,

		`$_at_least_once (cond=false) : *node.Root/*stmt.Do/*stmt.StmtList/*stmt.Expression/*node.SimpleVar`,
		`$_do_while_cond (cond=false) : *node.Root/*stmt.Do/*node.SimpleVar`,
	}
	linttest.RunFilterMatch(test, "pathTest")
}

func TestPathFuncScope(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function f() {
  echo $_inside_func;
  for (;;) {
    while (false) {
      return $_inside_loops;
    }
    return $_inside_loop;
  }
  return $_outside_loops;
}
`)
	test.Expect = []string{
		`$_inside_func (cond=false) : *stmt.Echo/*node.SimpleVar`,
		`$_inside_loops (cond=true) : *stmt.For/*stmt.StmtList/*stmt.While/*stmt.StmtList/*stmt.Return/*node.SimpleVar`,
		`$_inside_loop (cond=true) : *stmt.For/*stmt.StmtList/*stmt.Return/*node.SimpleVar`,
		`$_outside_loops (cond=false) : *stmt.Return/*node.SimpleVar`,
	}
	linttest.RunFilterMatch(test, "pathTest")
}

func TestPathMethodScope(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class C {
  public function f() {
    echo $_inside_func;
    for (;;) {
      while (false) {
        return $_inside_loops;
      }
      return $_inside_loop;
    }
    return $_outside_loops;
  }
}
`)
	test.Expect = []string{
		`$_inside_func (cond=false) : *stmt.Echo/*node.SimpleVar`,
		`$_inside_loops (cond=true) : *stmt.For/*stmt.StmtList/*stmt.While/*stmt.StmtList/*stmt.Return/*node.SimpleVar`,
		`$_inside_loop (cond=true) : *stmt.For/*stmt.StmtList/*stmt.Return/*node.SimpleVar`,
		`$_outside_loops (cond=false) : *stmt.Return/*node.SimpleVar`,
	}
	linttest.RunFilterMatch(test, "pathTest")
}

type pathTester struct {
	linter.BlockCheckerDefaults
	ctx *linter.BlockContext
}

func (b *pathTester) AfterEnterNode(w walker.Walkable) {
	if !meta.IsIndexingComplete() {
		return
	}

	x, ok := w.(*node.SimpleVar)
	if !ok || !strings.HasPrefix(x.Name, "_") {
		return
	}

	path := b.ctx.NodePath()
	b.ctx.Report(x, linter.LevelInformation, "pathTest", "$%s (cond=%v) : %s", x.Name, path.Conditional(), path.String())
}
