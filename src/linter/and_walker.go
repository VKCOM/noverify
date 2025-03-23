package linter

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
	"github.com/VKCOM/noverify/src/types"
	"github.com/VKCOM/noverify/src/utils"
)

// andWalker walks if conditions and adds isset/!empty/instanceof variables
// to the associated block walker.
//
// All variables defined by andWalker should be removed after
// if body is handled, this is why we collect varsToDelete.
type andWalker struct {
	b *blockWalker

	initialContext *blockContext
	// The context inside the if body if the condition is true.
	trueContext *blockContext
	// The context inside the else body if the condition is false.
	falseContext *blockContext

	varsToDelete []ir.Node

	path irutil.NodePath

	inNot bool
}

func (a *andWalker) exprType(n ir.Node) types.Map {
	return solver.ExprTypeCustom(a.b.ctx.sc, a.b.r.ctx.st, n, a.b.ctx.customTypes)
}

func (a *andWalker) exprTypeInContext(context *blockContext, n ir.Node) types.Map {
	return solver.ExprTypeCustom(context.sc, a.b.r.ctx.st, n, a.b.ctx.customTypes)
}

func (a *andWalker) EnterNode(w ir.Node) (res bool) {
	res = false

	switch n := w.(type) {
	case *ir.ParenExpr:
		return true

	case *ir.FunctionCallExpr:
		// If the absence of a function or method is being
		// checked, then nothing needs to be done.
		if a.inNot {
			return res
		}

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
		a.path.Push(n)
		n.Left.Walk(a)
		n.Right.Walk(a)
		a.path.Pop()

		a.runRules(w)
		return false
	case *ir.BooleanOrExpr:
		a.path.Push(n)
		n.Left.Walk(a)
		n.Right.Walk(a)
		a.path.Pop()

		a.runRules(w)
		return false

	case *ir.IssetExpr:
		for _, v := range n.Variables {
			varNode := utils.FindVarNode(v)
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

				// We need to traverse the variable here to check that
				// it exists, since this variable will be added to the
				// context later.
				a.b.handleVariable(varNode)

				var currentType types.Map
				if a.inNot {
					currentType = a.exprTypeInContext(a.trueContext, varNode)
				} else {
					currentType = a.exprTypeInContext(a.falseContext, varNode)
				}

				trueType := types.NewMap(className)
				falseType := currentType.Clone().Erase(className)

				if a.inNot {
					trueType, falseType = falseType, trueType
				}

				// If the variable has already been created, then we analyze the next instanceof.
				if (irutil.IsBoolAnd(a.path.Current()) || irutil.IsBoolOr(a.path.Current())) &&
					a.trueContext.sc.HaveImplicitVar(varNode) {

					if a.inNot {
						flags := meta.VarAlwaysDefined | meta.VarImplicit
						a.trueContext.sc.ReplaceVar(varNode, trueType, "instanceof true", flags)

						varInFalse, _ := a.falseContext.sc.GetVar(varNode)
						varInFalse.Type = varInFalse.Type.Append(falseType)
					} else {
						// The types in trueContext must be concatenated.
						varInTrue, _ := a.trueContext.sc.GetVar(varNode)
						varInTrue.Type = varInTrue.Type.Append(trueType)

						// And in falseContext, on the contrary, they are replaced,
						// since there are only types that are not in trueContext.
						flags := meta.VarAlwaysDefined | meta.VarImplicit
						a.falseContext.sc.ReplaceVar(varNode, falseType, "instanceof false", flags)
					}

				} else {
					flags := meta.VarAlwaysDefined | meta.VarImplicit
					a.trueContext.sc.ReplaceVar(varNode, trueType, "instanceof true", flags)
					a.falseContext.sc.ReplaceVar(varNode, falseType, "instanceof false", flags)
				}

			default:
				currentType := a.exprType(v)

				trueType := types.NewMap(className)
				falseType := currentType.Clone().Erase(className)

				if a.inNot {
					trueType, falseType = falseType, trueType
				}

				customTrueType := solver.CustomType{
					Node: n.Expr,
					Typ:  trueType,
				}
				customFalseType := solver.CustomType{
					Node: n.Expr,
					Typ:  falseType,
				}

				a.trueContext.customTypes = addCustomType(a.trueContext.customTypes, customTrueType)
				a.falseContext.customTypes = addCustomType(a.falseContext.customTypes, customFalseType)
			}
			// TODO: actually this needs to be present inside if body only
		}

	case *ir.NotIdenticalExpr:
		a.handleConditionSafety(n.Left, n.Right, false)
		a.handleConditionSafety(n.Right, n.Left, false)

	case *ir.IdenticalExpr:
		a.handleConditionSafety(n.Left, n.Right, true)
		a.handleConditionSafety(n.Right, n.Left, true)

	case *ir.BooleanNotExpr:
		a.inNot = true

		res = true
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
	return res
}

func (a *andWalker) handleConditionSafety(left ir.Node, right ir.Node, identical bool) {
	variable, ok := left.(*ir.SimpleVar)
	if !ok {
		return
	}

	constValue, ok := right.(*ir.ConstFetchExpr)
	if !ok || (constValue.Constant.Value != "false" && constValue.Constant.Value != "null") {
		return
	}

	// We need to traverse the variable here to check that
	// it exists, since this variable will be added to the
	// context later.
	a.b.handleVariable(variable)

	currentVar, isGotVar := a.trueContext.sc.GetVar(variable)
	if !isGotVar {
		return
	}

	var currentType types.Map
	if a.inNot {
		currentType = a.exprTypeInContext(a.trueContext, variable)
	} else {
		currentType = a.exprTypeInContext(a.falseContext, variable)
	}

	if constValue.Constant.Value == "false" || constValue.Constant.Value == "null" {
		clearType := currentType.Erase(constValue.Constant.Value)
		if identical {
			a.trueContext.sc.ReplaceVar(variable, currentType.Erase(clearType.String()), "type narrowing", currentVar.Flags)
			a.falseContext.sc.ReplaceVar(variable, clearType, "type narrowing", currentVar.Flags)
		} else {
			a.trueContext.sc.ReplaceVar(variable, clearType, "type narrowing", currentVar.Flags)
			a.falseContext.sc.ReplaceVar(variable, currentType.Erase(clearType.String()), "type narrowing", currentVar.Flags)
		}
		return
	}
}

func (a *andWalker) runRules(w ir.Node) {
	kind := ir.GetNodeKind(w)
	if a.b.r.anyRset != nil {
		a.b.r.runRules(w, a.b.ctx.sc, a.b.r.anyRset.RulesByKind[kind])
	} else if !a.b.rootLevel && a.b.r.localRset != nil {
		a.b.r.runRules(w, a.b.ctx.sc, a.b.r.localRset.RulesByKind[kind])
	}
}

func addCustomType(customTypes []solver.CustomType, typ solver.CustomType) []solver.CustomType {
	if len(customTypes) == 0 {
		return append(customTypes, typ)
	}

	// If a type has already been created for a node,
	// then the new type should replace it.
	if irutil.NodeEqual(customTypes[len(customTypes)-1].Node, typ.Node) {
		customTypes[len(customTypes)-1] = typ
		return customTypes
	}

	return append(customTypes, typ)
}

func (a *andWalker) LeaveNode(w ir.Node) {
	switch w.(type) {
	case *ir.BooleanNotExpr:
		a.inNot = false
	default:
	}
}
