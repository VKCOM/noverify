package phpgrep

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTracing(t *testing.T) {
	if !tracingEnabled {
		t.Skip("tracingEnabled is false; run tests with `-tags tracing`")
	}

	tests := []struct {
		pattern string
		input   string
		trace   []string
	}{
		// When outer matching fails, no other matching should be performed.
		{
			pattern: `if (0) $_`,
			input:   `if (1) { if (2) { if (3) {}}}`,
			trace: []string{
				"eqNode x=*ir.IfStmt y=*ir.IfStmt",
				" • eqNode x=*ir.Lnumber y=*ir.Lnumber",
			},
		},

		// When $_ is used for statement body it should quickly
		// match the body without checking the actual contents.
		{
			pattern: `if ($cond) $_`,
			input:   `if ($x && g()) { echo 1, 2, 3; }`,
			trace: []string{
				"eqNode x=*ir.IfStmt y=*ir.IfStmt",
				" • eqNode x=*ir.SimpleVar y=*ir.BooleanAndExpr",
				" • eqNode x=*ir.ExpressionStmt y=*ir.StmtList",
				" • eqNode x=<nil> y=<nil>",
			},
		},
		{
			pattern: `if ($cond) $_; else $_`,
			input:   `if (a(b(c()))) { 1; } else a(b(c()));`,
			trace: []string{
				"eqNode x=*ir.IfStmt y=*ir.IfStmt",
				" • eqNode x=*ir.SimpleVar y=*ir.FunctionCallExpr",
				" • eqNode x=*ir.ExpressionStmt y=*ir.StmtList",
				" • eqNode x=*ir.ElseStmt y=*ir.ElseStmt",
				" •  • eqNode x=*ir.ExpressionStmt y=*ir.ExpressionStmt",
			},
		},

		// The f(${"*"}) pattern is used for function call matching
		// that doesn't care about arguments.
		// It should not check arguments at all.
		{
			pattern: `f(${"*"})`,
			input:   `f(1, 2, 3, 4, g())`,
			trace: []string{
				"eqNode x=*ir.FunctionCallExpr y=*ir.FunctionCallExpr",
			},
		},

		// Match $_ once, then accept match as ${"*"} is the last node.
		{
			pattern: `f($_, ${"*"})`,
			input:   `f(1, 2, 3, 4, g())`,
			trace: []string{
				"eqNode x=*ir.FunctionCallExpr y=*ir.FunctionCallExpr",
				" • eqNode x=*ir.Argument y=*ir.Argument",
				" •  • eqNode x=*ir.SimpleVar y=*ir.Lnumber",
			},
		},

		// TODO: is it worthwhile to write a dedicated node slice function
		// for []*ir.Argument so we can avoid doing 1 extra eqNode per arg?
		// We need php-parser to store function args as *ir.Argument instead
		// of node.Node though.
		{
			pattern: `f(${"*"}, 5, ${"*"})`,
			input:   `f(1, 2, 3, 4, 5, 6, 7, 8)`,
			trace: []string{
				"eqNode x=*ir.FunctionCallExpr y=*ir.FunctionCallExpr",
				" • eqNode x=*ir.Argument y=*ir.Argument",
				" •  • eqNode x=*ir.Lnumber y=*ir.Lnumber",
				" • eqNode x=*ir.Argument y=*ir.Argument",
				" •  • eqNode x=*ir.Lnumber y=*ir.Lnumber",
				" • eqNode x=*ir.Argument y=*ir.Argument",
				" •  • eqNode x=*ir.Lnumber y=*ir.Lnumber",
				" • eqNode x=*ir.Argument y=*ir.Argument",
				" •  • eqNode x=*ir.Lnumber y=*ir.Lnumber",
				" • eqNode x=*ir.Argument y=*ir.Argument",
				" •  • eqNode x=*ir.Lnumber y=*ir.Lnumber",
			},
		},

		// Mismatching argument lists length should cause matching to stop
		// before we enter another eqNode().
		//
		// TODO: compile patterns with known args length into
		// something that checks arguments count before matching the
		// function name?
		{
			pattern: `sizeof($_)`,
			input:   `sizeof()`,
			trace: []string{
				"eqNode x=*ir.FunctionCallExpr y=*ir.FunctionCallExpr",
			},
		},
		{
			pattern: `sizeof()`,
			input:   `sizeof($_)`,
			trace: []string{
				"eqNode x=*ir.FunctionCallExpr y=*ir.FunctionCallExpr",
			},
		},

		// If operator mismatches, operands are not processed.
		{
			pattern: `$x + $y`,
			input:   `$x - $y`,
			trace: []string{
				"eqNode x=*ir.PlusExpr y=*ir.MinusExpr",
			},
		},

		{
			pattern: `$x + $y`,
			input:   `1 + 2`,
			trace: []string{
				"eqNode x=*ir.PlusExpr y=*ir.PlusExpr",
				" • eqNode x=*ir.SimpleVar y=*ir.Lnumber",
				" • eqNode x=*ir.SimpleVar y=*ir.Lnumber",
			},
		},

		{
			pattern: `[${"*"}, 123, ${"*"}]`,
			input:   `[A, B,  C, D]`,
			trace: []string{
				"eqNode x=*ir.ArrayExpr y=*ir.ArrayExpr",
				" • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  • eqNode x=*ir.Lnumber y=*ir.ConstFetchExpr",
				" • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  • eqNode x=*ir.Lnumber y=*ir.ConstFetchExpr",
				" • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  • eqNode x=*ir.Lnumber y=*ir.ConstFetchExpr",
				" • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  • eqNode x=*ir.Lnumber y=*ir.ConstFetchExpr",
			},
		},

		// TODO: ${"*"} could theoretically give up after trying to
		// match `3` as there are less nodes left in the slice than
		// fixed-length nodes left in the pattern.
		{
			pattern: `{ ${"*"}; a(); b(); c(); }`,
			input:   `{ 1; 2; 3; 4; 5; }`,
			trace: []string{
				"eqNode x=*ir.StmtList y=*ir.StmtList",
				" • eqNode x=*ir.ExpressionStmt y=*ir.ExpressionStmt",
				" •  • eqNode x=*ir.FunctionCallExpr y=*ir.Lnumber",
				" • eqNode x=*ir.ExpressionStmt y=*ir.ExpressionStmt",
				" •  • eqNode x=*ir.FunctionCallExpr y=*ir.Lnumber",
				" • eqNode x=*ir.ExpressionStmt y=*ir.ExpressionStmt",
				" •  • eqNode x=*ir.FunctionCallExpr y=*ir.Lnumber",
				" • eqNode x=*ir.ExpressionStmt y=*ir.ExpressionStmt",
				" •  • eqNode x=*ir.FunctionCallExpr y=*ir.Lnumber",
				" • eqNode x=*ir.ExpressionStmt y=*ir.ExpressionStmt",
				" •  • eqNode x=*ir.FunctionCallExpr y=*ir.Lnumber",
			},
		},

		// This pattern cuts 2 ${"*"} after it matches ${"str"}.
		{
			pattern: `f([${"*"}, $_ => [${"*"}, ${"str"}, ${"*"}], ${"*"}])`,
			input:   `f([$x, A => [B, C, 'bingo'], $y])`,
			trace: []string{
				"eqNode x=*ir.FunctionCallExpr y=*ir.FunctionCallExpr",
				" • eqNode x=*ir.Argument y=*ir.Argument",
				" •  • eqNode x=*ir.ArrayExpr y=*ir.ArrayExpr",
				" •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  •  • eqNode x=*ir.SimpleVar y=*ir.ConstFetchExpr",
				" •  •  •  • eqNode x=*ir.ArrayExpr y=*ir.ArrayExpr",
				" •  •  •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  •  •  •  • eqNode x=*ir.Var y=*ir.ConstFetchExpr",
				" •  •  •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  •  •  •  • eqNode x=*ir.Var y=*ir.ConstFetchExpr",
				" •  •  •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  •  •  •  • eqNode x=*ir.Var y=*ir.String",
			},
		},

		{
			pattern: `f([${"*"}, $_ => [${"*"}, ${"str"}, ${"*"}], ${"*"}])`,
			input:   `f([$x, A => [B, C, D], $y, $z])`,
			trace: []string{
				"eqNode x=*ir.FunctionCallExpr y=*ir.FunctionCallExpr",
				" • eqNode x=*ir.Argument y=*ir.Argument",
				" •  • eqNode x=*ir.ArrayExpr y=*ir.ArrayExpr",
				" •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  •  • eqNode x=*ir.SimpleVar y=*ir.ConstFetchExpr",
				" •  •  •  • eqNode x=*ir.ArrayExpr y=*ir.ArrayExpr",
				" •  •  •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  •  •  •  • eqNode x=*ir.Var y=*ir.ConstFetchExpr",
				" •  •  •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  •  •  •  • eqNode x=*ir.Var y=*ir.ConstFetchExpr",
				" •  •  •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  •  •  •  • eqNode x=*ir.Var y=*ir.ConstFetchExpr",
				" •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
				" •  •  • eqNode x=*ir.ArrayItemExpr y=*ir.ArrayItemExpr",
			},
		},
	}

	var c Compiler
	for _, test := range tests {
		var buf bytes.Buffer
		matcher := mustCompile(t, &c, test.pattern)
		matcher.m.tracingBuf = &buf
		matcher.Match(mustParse(t, test.input))
		trace := strings.Split(buf.String(), "\n")
		trace = trace[:len(trace)-1] // Drop ""
		if diff := cmp.Diff(trace, test.trace); diff != "" {
			t.Errorf("`%s` (-have +want):\n%s", test.pattern, diff)
		}
	}
}
