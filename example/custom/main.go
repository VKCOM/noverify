package main

// This is an example of adding of custom rules

import (
	"flag"
	"log"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/solver"
)

var customFlag = flag.String("custom-flag", "", "An example of the additional linter flag")

func addCheckers(config *linter.Config) {
	config.Checkers.AddBlockChecker(func(ctx *linter.BlockContext) linter.BlockChecker {
		return &block{ctx: ctx}
	})
	config.Checkers.DeclareChecker(linter.CheckerInfo{
		Name:    "exampleStrictCmp",
		Default: true,
		Comment: "Report not-strict-enough comparisons.",
	})
}

func main() {
	log.SetFlags(log.Flags() | log.Ltime)

	config := linter.NewConfig()
	addCheckers(config)

	// Config argument can be nil to use "all default" behavior.
	cmd.Main(&cmd.MainConfig{
		AfterFlagParse: useCustomFlags,
		LinterConfig:   config,
	})
}

func useCustomFlags(env cmd.InitEnvironment) {
	if *customFlag != "" {
		log.Println("custom flag value:", *customFlag)
	}
}

type block struct {
	linter.BlockCheckerDefaults
	ctx *linter.BlockContext
}

func isString(ctx *linter.BlockContext, n ir.Node) bool {
	_, ok := n.(*ir.String)
	if ok {
		return true
	}

	return solver.ExprType(ctx.Scope(), ctx.ClassParseState(), n).Is("string")
}

func (b *block) BeforeEnterNode(n ir.Node) {
	switch n := n.(type) {
	case *ir.FunctionCallExpr:
		b.handleFunctionCall(n)
	case *ir.EqualExpr:
		if isString(b.ctx, n.Left) || isString(b.ctx, n.Right) {
			b.ctx.Report(n, linter.LevelWarning, "exampleStrictCmp", "Strings must be compared using '===' operator")
		}
	case *ir.NotEqualExpr:
		if isString(b.ctx, n.Left) || isString(b.ctx, n.Right) {
			b.ctx.Report(n, linter.LevelWarning, "exampleStrictCmp", "Strings must be compared using '!==' operator")
		}
	}
}

func (b *block) handleFunctionCall(e *ir.FunctionCallExpr) {
	nm, ok := e.Function.(*ir.Name)
	if !ok {
		return
	}

	if nm.Value == `in_array` {
		b.handleInArrayCall(e)
		return
	}
}

func (b *block) handleInArrayCall(e *ir.FunctionCallExpr) {
	if len(e.Args) != 2 {
		return
	}

	arg := e.Arg(0)
	if !isString(b.ctx, arg.Expr) {
		return
	}

	b.ctx.Report(e, linter.LevelWarning, "exampleStrictCmp", "3rd argument of in_array must be true when comparing strings")
}
