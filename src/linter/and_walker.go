package linter

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/solver"
)

// andWalker walks if conditions and adds isset/!empty/instanceof variables
// to the associated block walker.
//
// All variables defined by andWalker should be removed after
// if body is handled, this is why we collect varsToDelete.
type andWalker struct {
	b *BlockWalker

	varsToDelete []node.Node
}

func (a *andWalker) EnterNode(w walker.Walkable) (res bool) {
	switch n := w.(type) {
	case *expr.FunctionCall:
		args := n.ArgumentList.Arguments
		nm, ok := n.Function.(*name.Name)
		if !ok {
			break
		}
		switch {
		case len(args) == 2 && meta.NameEquals(nm, `method_exists`):
			obj := args[0].(*node.Argument).Expr
			methodName := args[1].(*node.Argument).Expr
			lit, ok := methodName.(*scalar.String)
			if ok {
				a.b.ctx.addCustomMethod(obj, unquote(lit.Value))
			}
		case len(args) == 1 && meta.NameEquals(nm, `function_exists`):
			functionName := args[0].(*node.Argument).Expr
			lit, ok := functionName.(*scalar.String)
			if ok {
				a.b.ctx.addCustomFunction(unquote(lit.Value))
			}
		}

	case *binary.BooleanAnd:
		return true

	case *expr.Isset:
		for _, v := range n.Variables {
			varNode := findVarNode(v)
			if varNode == nil {
				continue
			}
			if a.b.ctx.sc.HaveVar(varNode) {
				continue
			}

			switch v := varNode.(type) {
			case *node.SimpleVar:
				a.b.addVar(v, meta.NewTypesMap("isset_$"+v.Name), "isset", meta.VarAlwaysDefined)
				a.varsToDelete = append(a.varsToDelete, v)
			case *node.Var:
				a.b.handleVariable(v.Expr)
				vv, ok := v.Expr.(*node.SimpleVar)
				if !ok {
					continue
				}
				a.b.addVar(v, meta.NewTypesMap("isset_$$"+vv.Name), "isset", meta.VarAlwaysDefined)
				a.varsToDelete = append(a.varsToDelete, v)
			}
		}

	case *expr.InstanceOf:
		if className, ok := solver.GetClassName(a.b.r.ctx.st, n.Class); ok {
			switch v := n.Expr.(type) {
			case *node.Var, *node.SimpleVar:
				a.b.ctx.sc.AddVar(v, meta.NewTypesMap(className), "instanceof", 0)
			default:
				a.b.ctx.customTypes = append(a.b.ctx.customTypes, solver.CustomType{
					Node: n.Expr,
					Typ:  meta.NewTypesMap(className),
				})
			}
			// TODO: actually this needs to be present inside if body only
		}

	case *expr.BooleanNot:
		// TODO: consolidate with issets handling?
		// Probably could collect *expr.Variable instead of
		// isset and empty nodes and handle them in a single loop.

		// !empty($x) implies that isset($x) would return true.
		empty, ok := n.Expr.(*expr.Empty)
		if !ok {
			break
		}
		v, ok := empty.Expr.(*node.SimpleVar)
		if !ok {
			break
		}
		if a.b.ctx.sc.HaveVar(v) {
			break
		}
		a.b.addVar(v, meta.NewTypesMap("isset_$"+v.Name), "!empty", meta.VarAlwaysDefined)
		a.varsToDelete = append(a.varsToDelete, v)
	}

	w.Walk(a.b)
	return false
}

func (a *andWalker) LeaveNode(w walker.Walkable) {}
