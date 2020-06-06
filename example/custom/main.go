package main

// This is an example of adding of custom rules

import (
	"flag"
	"log"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/solver"
)

func init() {
	linter.RegisterBlockChecker(func(ctx *linter.BlockContext) linter.BlockChecker { return &block{ctx: ctx} })

	linter.DeclareCheck(linter.CheckInfo{
		Name:    "strictCmp",
		Default: true,
		Comment: "Report not-strict-enough comparisons.",
	})
}

var customFlag = flag.String("custom-flag", "", "An example of the additional linter flag")

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)

	// Config argument can be nil to use "all default" behavior.
	cmd.Main(&cmd.MainConfig{
		AfterFlagParse: useCustomFlags,
	})
}

func useCustomFlags() {
	if *customFlag != "" {
		log.Println("custom flag value:", *customFlag)
	}
}

type block struct {
	linter.BlockCheckerDefaults
	ctx *linter.BlockContext
}

func isString(ctx *linter.BlockContext, n node.Node) bool {
	_, ok := n.(*scalar.String)
	if ok {
		return true
	}

	return solver.ExprType(ctx.Scope(), ctx.ClassParseState(), n).Is("string")
}

func (b *block) BeforeEnterNode(w walker.Walkable) {
	switch n := w.(type) {
	case *expr.FunctionCall:
		b.handleFunctionCall(n)
	case *binary.Equal:
		if isString(b.ctx, n.Left) || isString(b.ctx, n.Right) {
			b.ctx.Report(n, linter.LevelWarning, "strictCmp", "Strings must be compared using '===' operator")
		}
	case *binary.NotEqual:
		if isString(b.ctx, n.Left) || isString(b.ctx, n.Right) {
			b.ctx.Report(n, linter.LevelWarning, "strictCmp", "Strings must be compared using '!==' operator")
		}
	}
}

func (b *block) handleFunctionCall(e *expr.FunctionCall) {
	nm, ok := e.Function.(*name.Name)
	if !ok {
		return
	}

	if meta.NameEquals(nm, `in_array`) {
		b.handleInArrayCall(e)
		return
	}
}

func (b *block) handleInArrayCall(e *expr.FunctionCall) {
	if len(e.ArgumentList.Arguments) != 2 {
		return
	}

	arg := e.ArgumentList.Arguments[0].(*node.Argument)
	if !isString(b.ctx, arg.Expr) {
		return
	}

	b.ctx.Report(e, linter.LevelWarning, "strictCmp", "3rd argument of in_array must be true when comparing strings")
}
