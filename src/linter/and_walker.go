package linter

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
)

type varToReplace struct {
	Node ir.Node
	Type types.Map
}

// andWalker walks if conditions and adds isset/!empty/instanceof variables
// to the associated block walker.
//
// All variables defined by andWalker should be removed after
// if body is handled, this is why we collect varsToDelete.
type andWalker struct {
	b *blockWalker

	// The context inside the if body if the condition is true.
	trueContext *blockContext
	// The context inside the else body if the condition is false.
	falseContext *blockContext

	varsToDelete  []ir.Node
	varsToReplace []varToReplace
}

func (a *andWalker) exprType(n ir.Node) types.Map {
	return solver.ExprType(a.b.ctx.sc, a.b.r.ctx.st, n)
}

func (a *andWalker) EnterNode(w ir.Node) (res bool) {
	switch n := w.(type) {
	case *ir.FunctionCallExpr:
		nm, ok := n.Function.(*ir.Name)
		if !ok {
			break
		}
		switch {
		case len(n.Args) == 2 && nm.Value == `method_exists`:
			obj := n.Arg(0).Expr
			methodName := n.Arg(1).Expr
			lit, ok := methodName.(*ir.String)
			if ok {
				a.b.ctx.addCustomMethod(obj, lit.Value)
			}
		case len(n.Args) == 1 && nm.Value == `function_exists`:
			functionName := n.Arg(0).Expr
			lit, ok := functionName.(*ir.String)
			if ok {
				a.b.ctx.addCustomFunction(lit.Value)
			}
		}

	case *ir.BooleanAndExpr:
		return true

	case *ir.IssetExpr:
		for _, v := range n.Variables {
			varNode := findVarNode(v)
			if varNode == nil {
				continue
			}
			if a.b.ctx.sc.HaveVar(varNode) {
				continue
			}

			switch v := varNode.(type) {
			case *ir.SimpleVar:
				a.b.addVar(v, types.NewMap("isset_$"+v.Name), "isset", meta.VarAlwaysDefined)
				a.varsToDelete = append(a.varsToDelete, v)
			case *ir.Var:
				a.b.handleVariable(v.Expr)
				vv, ok := v.Expr.(*ir.SimpleVar)
				if !ok {
					continue
				}
				a.b.addVar(v, types.NewMap("isset_$$"+vv.Name), "isset", meta.VarAlwaysDefined)
				a.varsToDelete = append(a.varsToDelete, v)
			}
		}

	case *ir.InstanceOfExpr:
		if className, ok := solver.GetClassName(a.b.r.ctx.st, n.Class); ok {
			switch v := n.Expr.(type) {
			case *ir.Var, *ir.SimpleVar:
				varNode := v
				a.b.handleVariable(varNode)

				currentType := a.exprType(varNode)
				trueType := types.NewMap(className)
				falseType := currentType.Clone().Erase(className)

				a.trueContext.sc.ReplaceVar(varNode, trueType, "instanceof true", meta.VarAlwaysDefined)
				a.falseContext.sc.ReplaceVar(varNode, falseType, "instanceof false", meta.VarAlwaysDefined)

				a.varsToReplace = append(a.varsToReplace, varToReplace{
					Node: varNode,
					Type: currentType,
				})

			default:
				a.b.ctx.customTypes = append(a.b.ctx.customTypes, solver.CustomType{
					Node: n.Expr,
					Typ:  types.NewMap(className),
				})
			}
			// TODO: actually this needs to be present inside if body only
		}

	case *ir.BooleanNotExpr:
		// TODO: consolidate with issets handling?
		// Probably could collect *expr.Variable instead of
		// isset and empty nodes and handle them in a single loop.

		// !empty($x) implies that isset($x) would return true.
		empty, ok := n.Expr.(*ir.EmptyExpr)
		if !ok {
			break
		}
		v, ok := empty.Expr.(*ir.SimpleVar)
		if !ok {
			break
		}
		if a.b.ctx.sc.HaveVar(v) {
			break
		}
		a.b.addVar(v, types.NewMap("isset_$"+v.Name), "!empty", meta.VarAlwaysDefined)
		a.varsToDelete = append(a.varsToDelete, v)
	}

	w.Walk(a.b)
	return false
}

func (a *andWalker) LeaveNode(w ir.Node) {}
