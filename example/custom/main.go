package main

// This is an example of adding of custom rules

import (
	"log"

	"github.com/VKCOM/noverify/src/cmd"
	"github.com/VKCOM/noverify/src/linter"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/walker"
)

func main() {
	log.SetFlags(log.Flags() | log.Lmicroseconds)
	linter.RegisterBlockChecker(func(ctx linter.BlockContext) linter.BlockChecker { return &block{ctx: ctx} })
	cmd.Main()
}

type block struct {
	ctx linter.BlockContext
}

func isString(ctx linter.BlockContext, n node.Node) bool {
	_, ok := n.(*scalar.String)
	if ok {
		return true
	}

	return solver.ExprType(ctx.Scope(), ctx.ClassParseState(), n).IsString()
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
	if len(e.Arguments) != 2 {
		return
	}

	arg, ok := e.Arguments[0].(*node.Argument)
	if !ok {
		return
	}

	if !isString(b.ctx, arg.Expr) {
		return
	}

	b.ctx.Report(e, linter.LevelWarning, "strictCmp", "3rd argument of in_array must be true when comparing strings")
}

func (b *block) AfterEnterNode(w walker.Walkable)  {}
func (b *block) BeforeLeaveNode(w walker.Walkable) {}
func (b *block) AfterLeaveNode(w walker.Walkable)  {}
