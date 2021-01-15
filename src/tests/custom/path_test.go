package custom

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/linttest"
	"github.com/VKCOM/noverify/src/meta"
)

func init() {
	linter.RegisterBlockChecker(func(ctx *linter.BlockContext) linter.BlockChecker {
		return &pathTester{ctx: ctx}
	})
}

func runPathTest(t *testing.T, suite *linttest.Suite) {
	t.Helper()
	suite.IgnoreUndeclaredChecks = true
	linttest.RunFilterMatch(suite, "pathTest")
}

func TestPathTryCatch(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
try {
  echo $_try;
} catch (Exception $e) {
  echo $_catch;
}
`)
	test.Expect = []string{
		`$_try (cond=false) : *ir.Root/*ir.TryStmt/*ir.EchoStmt/*ir.SimpleVar`,
		`$_catch (cond=true) : *ir.Root/*ir.TryStmt/*ir.CatchStmt/*ir.EchoStmt/*ir.SimpleVar`,
	}
	runPathTest(t, test)
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
		`$_if (cond=true) : *ir.Root/*ir.IfStmt/*ir.StmtList/*ir.EchoStmt/*ir.SimpleVar`,
		`$_elseif1 (cond=true) : *ir.Root/*ir.IfStmt/*ir.StmtList/*ir.EchoStmt/*ir.SimpleVar`,
		`$_elseif2 (cond=true) : *ir.Root/*ir.IfStmt/*ir.StmtList/*ir.ExpressionStmt/*ir.SimpleVar`,
		`$_else (cond=true) : *ir.Root/*ir.IfStmt/*ir.ElseStmt/*ir.StmtList/*ir.EchoStmt/*ir.SimpleVar`,
	}
	runPathTest(t, test)
}

func TestPathArgument(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
f($_arg1, [$_arg2], $a + $_arg3);
`)
	test.Expect = []string{
		`$_arg1 (cond=false) : *ir.Root/*ir.ExpressionStmt/*ir.FunctionCallExpr/*ir.Argument/*ir.SimpleVar`,
		`$_arg2 (cond=false) : *ir.Root/*ir.ExpressionStmt/*ir.FunctionCallExpr/*ir.Argument/*ir.ArrayExpr/*ir.SimpleVar`,
		`$_arg3 (cond=false) : *ir.Root/*ir.ExpressionStmt/*ir.FunctionCallExpr/*ir.Argument/*ir.PlusExpr/*ir.SimpleVar`,
	}
	runPathTest(t, test)
}

func TestPathTernary(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
echo $_ternary_cond ? $_ternary_true : $_ternary_false;
`)
	test.Expect = []string{
		`$_ternary_cond (cond=true) : *ir.Root/*ir.EchoStmt/*ir.TernaryExpr/*ir.SimpleVar`,
		`$_ternary_true (cond=true) : *ir.Root/*ir.EchoStmt/*ir.TernaryExpr/*ir.SimpleVar`,
		`$_ternary_false (cond=true) : *ir.Root/*ir.EchoStmt/*ir.TernaryExpr/*ir.SimpleVar`,
	}
	runPathTest(t, test)
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
		`$_cond (cond=true) : *ir.Root/*ir.IfStmt/*ir.SimpleVar`,
		`$_x1 (cond=true) : *ir.Root/*ir.IfStmt/*ir.StmtList/*ir.EchoStmt/*ir.SimpleVar`,
		`$_case (cond=true) : *ir.Root/*ir.IfStmt/*ir.StmtList/*ir.SwitchStmt/*ir.SimpleVar`,
		`$_x2 (cond=true) : *ir.Root/*ir.IfStmt/*ir.StmtList/*ir.SwitchStmt/*ir.EchoStmt/*ir.SimpleVar`,

		`$_at_least_once (cond=false) : *ir.Root/*ir.DoStmt/*ir.StmtList/*ir.ExpressionStmt/*ir.SimpleVar`,
		`$_do_while_cond (cond=false) : *ir.Root/*ir.DoStmt/*ir.SimpleVar`,
	}
	runPathTest(t, test)
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
		`$_inside_func (cond=false) : *ir.EchoStmt/*ir.SimpleVar`,
		`$_inside_loops (cond=true) : *ir.ForStmt/*ir.StmtList/*ir.WhileStmt/*ir.StmtList/*ir.ReturnStmt/*ir.SimpleVar`,
		`$_inside_loop (cond=true) : *ir.ForStmt/*ir.StmtList/*ir.ReturnStmt/*ir.SimpleVar`,
		`$_outside_loops (cond=false) : *ir.ReturnStmt/*ir.SimpleVar`,
	}
	runPathTest(t, test)
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
		`$_inside_func (cond=false) : *ir.EchoStmt/*ir.SimpleVar`,
		`$_inside_loops (cond=true) : *ir.ForStmt/*ir.StmtList/*ir.WhileStmt/*ir.StmtList/*ir.ReturnStmt/*ir.SimpleVar`,
		`$_inside_loop (cond=true) : *ir.ForStmt/*ir.StmtList/*ir.ReturnStmt/*ir.SimpleVar`,
		`$_outside_loops (cond=false) : *ir.ReturnStmt/*ir.SimpleVar`,
	}
	runPathTest(t, test)
}

type pathTester struct {
	linter.BlockCheckerDefaults
	ctx *linter.BlockContext
}

func (b *pathTester) AfterEnterNode(n ir.Node) {
	if !meta.IsIndexingComplete() {
		return
	}

	x, ok := n.(*ir.SimpleVar)
	if !ok || !strings.HasPrefix(x.Name, "_") {
		return
	}

	path := b.ctx.NodePath()
	b.ctx.Report(x, linter.LevelInfo, "pathTest", "$%s (cond=%v) : %s", x.Name, path.Conditional(), path.String())
}
