package linter

import (
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/solver"
)

// assignWalker handles assignments by pushing all defined variables
// to the associated block scope.
type assignWalker struct {
	b *BlockWalker
}

func (a *assignWalker) EnterNode(w walker.Walkable) (res bool) {
	b := a.b
	switch n := w.(type) {
	case *assign.Assign:
		switch v := n.Variable.(type) {
		case *node.Var, *node.SimpleVar:
			b.ctx.sc.ReplaceVar(v, solver.ExprTypeLocal(b.ctx.sc, b.r.st, n.Expression), "assign", true)
		}
	}
	return true
}

func (a *assignWalker) LeaveNode(w walker.Walkable) {}
