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
				"eqNode x=*stmt.If y=*stmt.If",
				" • eqNode x=*scalar.Lnumber y=*scalar.Lnumber",
			},
		},

		// When $_ is used for statement body it should quickly
		// match the body without checking the actual contents.
		{
			pattern: `if ($cond) $_`,
			input:   `if ($x && g()) { echo 1, 2, 3; }`,
			trace: []string{
				"eqNode x=*stmt.If y=*stmt.If",
				" • eqNode x=*node.SimpleVar y=*binary.BooleanAnd",
				" • eqNode x=*stmt.Expression y=*stmt.StmtList",
				" • eqNode x=<nil> y=<nil>",
			},
		},
		{
			pattern: `if ($cond) $_; else $_`,
			input:   `if (a(b(c()))) { 1; } else a(b(c()));`,
			trace: []string{
				"eqNode x=*stmt.If y=*stmt.If",
				" • eqNode x=*node.SimpleVar y=*expr.FunctionCall",
				" • eqNode x=*stmt.Expression y=*stmt.StmtList",
				" • eqNode x=*stmt.Else y=*stmt.Else",
				" •  • eqNode x=*stmt.Expression y=*stmt.Expression",
			},
		},

		// The f(${"*"}) pattern is used for function call matching
		// that doesn't care about arguments.
		// It should not check arguments at all.
		{
			pattern: `f(${"*"})`,
			input:   `f(1, 2, 3, 4, g())`,
			trace: []string{
				"eqNode x=*expr.FunctionCall y=*expr.FunctionCall",
				" • eqNode x=*name.Name y=*name.Name",
			},
		},

		// Match $_ once, then accept match as ${"*"} is the last node.
		{
			pattern: `f($_, ${"*"})`,
			input:   `f(1, 2, 3, 4, g())`,
			trace: []string{
				"eqNode x=*expr.FunctionCall y=*expr.FunctionCall",
				" • eqNode x=*name.Name y=*name.Name",
				" • eqNode x=*node.Argument y=*node.Argument",
				" •  • eqNode x=*node.SimpleVar y=*scalar.Lnumber",
			},
		},

		// TODO: is it worthwhile to write a dedicated node slice function
		// for []*node.Argument so we can avoid doing 1 extra eqNode per arg?
		// We need php-parser to store function args as *node.Argument instead
		// of node.Node though.
		{
			pattern: `f(${"*"}, 5, ${"*"})`,
			input:   `f(1, 2, 3, 4, 5, 6, 7, 8)`,
			trace: []string{
				"eqNode x=*expr.FunctionCall y=*expr.FunctionCall",
				" • eqNode x=*name.Name y=*name.Name",
				" • eqNode x=*node.Argument y=*node.Argument",
				" •  • eqNode x=*scalar.Lnumber y=*scalar.Lnumber",
				" • eqNode x=*node.Argument y=*node.Argument",
				" •  • eqNode x=*scalar.Lnumber y=*scalar.Lnumber",
				" • eqNode x=*node.Argument y=*node.Argument",
				" •  • eqNode x=*scalar.Lnumber y=*scalar.Lnumber",
				" • eqNode x=*node.Argument y=*node.Argument",
				" •  • eqNode x=*scalar.Lnumber y=*scalar.Lnumber",
				" • eqNode x=*node.Argument y=*node.Argument",
				" •  • eqNode x=*scalar.Lnumber y=*scalar.Lnumber",
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
				"eqNode x=*expr.FunctionCall y=*expr.FunctionCall",
				" • eqNode x=*name.Name y=*name.Name",
			},
		},
		{
			pattern: `sizeof()`,
			input:   `sizeof($_)`,
			trace: []string{
				"eqNode x=*expr.FunctionCall y=*expr.FunctionCall",
				" • eqNode x=*name.Name y=*name.Name",
			},
		},

		// If operator mismatches, operands are not processed.
		{
			pattern: `$x + $y`,
			input:   `$x - $y`,
			trace: []string{
				"eqNode x=*binary.Plus y=*binary.Minus",
			},
		},

		{
			pattern: `$x + $y`,
			input:   `1 + 2`,
			trace: []string{
				"eqNode x=*binary.Plus y=*binary.Plus",
				" • eqNode x=*node.SimpleVar y=*scalar.Lnumber",
				" • eqNode x=*node.SimpleVar y=*scalar.Lnumber",
			},
		},

		{
			pattern: `[${"*"}, 123, ${"*"}]`,
			input:   `[A, B,  C, D]`,
			trace: []string{
				"eqNode x=*expr.Array y=*expr.Array",
				" • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  • eqNode x=*scalar.Lnumber y=*expr.ConstFetch",
				" • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  • eqNode x=*scalar.Lnumber y=*expr.ConstFetch",
				" • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  • eqNode x=*scalar.Lnumber y=*expr.ConstFetch",
				" • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  • eqNode x=*scalar.Lnumber y=*expr.ConstFetch",
			},
		},

		// TODO: ${"*"} could theoretically give up after trying to
		// match `3` as there are less nodes left in the slice than
		// fixed-length nodes left in the pattern.
		{
			pattern: `{ ${"*"}; a(); b(); c(); }`,
			input:   `{ 1; 2; 3; 4; 5; }`,
			trace: []string{
				"eqNode x=*stmt.StmtList y=*stmt.StmtList",
				" • eqNode x=*stmt.Expression y=*stmt.Expression",
				" •  • eqNode x=*expr.FunctionCall y=*scalar.Lnumber",
				" • eqNode x=*stmt.Expression y=*stmt.Expression",
				" •  • eqNode x=*expr.FunctionCall y=*scalar.Lnumber",
				" • eqNode x=*stmt.Expression y=*stmt.Expression",
				" •  • eqNode x=*expr.FunctionCall y=*scalar.Lnumber",
				" • eqNode x=*stmt.Expression y=*stmt.Expression",
				" •  • eqNode x=*expr.FunctionCall y=*scalar.Lnumber",
				" • eqNode x=*stmt.Expression y=*stmt.Expression",
				" •  • eqNode x=*expr.FunctionCall y=*scalar.Lnumber",
			},
		},

		// This pattern cuts 2 ${"*"} after it matches ${"str"}.
		{
			pattern: `f([${"*"}, $_ => [${"*"}, ${"str"}, ${"*"}], ${"*"}])`,
			input:   `f([$x, A => [B, C, 'bingo'], $y])`,
			trace: []string{
				"eqNode x=*expr.FunctionCall y=*expr.FunctionCall",
				" • eqNode x=*name.Name y=*name.Name",
				" • eqNode x=*node.Argument y=*node.Argument",
				" •  • eqNode x=*expr.Array y=*expr.Array",
				" •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  •  • eqNode x=*node.SimpleVar y=*expr.ConstFetch",
				" •  •  •  • eqNode x=*expr.Array y=*expr.Array",
				" •  •  •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  •  •  •  • eqNode x=*node.Var y=*expr.ConstFetch",
				" •  •  •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  •  •  •  • eqNode x=*node.Var y=*expr.ConstFetch",
				" •  •  •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  •  •  •  • eqNode x=*node.Var y=*scalar.String",
			},
		},

		{
			pattern: `f([${"*"}, $_ => [${"*"}, ${"str"}, ${"*"}], ${"*"}])`,
			input:   `f([$x, A => [B, C, D], $y, $z])`,
			trace: []string{
				"eqNode x=*expr.FunctionCall y=*expr.FunctionCall",
				" • eqNode x=*name.Name y=*name.Name",
				" • eqNode x=*node.Argument y=*node.Argument",
				" •  • eqNode x=*expr.Array y=*expr.Array",
				" •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  •  • eqNode x=*node.SimpleVar y=*expr.ConstFetch",
				" •  •  •  • eqNode x=*expr.Array y=*expr.Array",
				" •  •  •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  •  •  •  • eqNode x=*node.Var y=*expr.ConstFetch",
				" •  •  •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  •  •  •  • eqNode x=*node.Var y=*expr.ConstFetch",
				" •  •  •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  •  •  •  • eqNode x=*node.Var y=*expr.ConstFetch",
				" •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
				" •  •  • eqNode x=*expr.ArrayItem y=*expr.ArrayItem",
			},
		},
	}

	for _, test := range tests {
		var buf bytes.Buffer
		matcher := mustCompile(t, test.pattern)
		matcher.m.tracingBuf = &buf
		matcher.Match(mustParse(t, test.input))
		trace := strings.Split(buf.String(), "\n")
		trace = trace[:len(trace)-1] // Drop ""
		if diff := cmp.Diff(trace, test.trace); diff != "" {
			t.Errorf("`%s` (-have +want):\n%s", test.pattern, diff)
		}
	}
}
