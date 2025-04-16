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

	case *ir.SimpleVar:
		a.handleVariableCondition(n)

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

		switch {
		case nm.Value == `is_int`:
			a.handleTypeCheckCondition("int", n.Args)
		case nm.Value == `is_float`:
			a.handleTypeCheckCondition("float", n.Args)
		case nm.Value == `is_string`:
			a.handleTypeCheckCondition("string", n.Args)
		case nm.Value == `is_object`:
			a.handleTypeCheckCondition("object", n.Args)
		case nm.Value == `is_array`:
			a.handleTypeCheckCondition("array", n.Args)
		case nm.Value == `is_null`:
			a.handleTypeCheckCondition("null", n.Args)
		case nm.Value == `is_resource`:
			a.handleTypeCheckCondition("resource", n.Args)
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

func (a *andWalker) handleVariableCondition(variable *ir.SimpleVar) {
	if !a.b.ctx.sc.HaveVar(variable) {
		return
	}

	currentType := a.exprType(variable) // nolint:staticcheck
	if a.inNot {
		currentType = a.exprTypeInContext(a.trueContext, variable)
	} else {
		currentType = a.exprTypeInContext(a.falseContext, variable)
	}

	var trueType, falseType types.Map

	// First, handle "null": if currentType contains "null", then in the true branch we remove it,
	// and in the false branch we narrow to "null"
	if currentType.Contains("null") {
		trueType = currentType.Clone().Erase("null")
		falseType = types.NewMap("null")
	} else {
		trueType = currentType.Clone()
		falseType = currentType.Clone()
	}

	// Next, handle booleans
	// If currentType contains any boolean-related literal ("bool", "true", "false"),
	// then we want to narrow them:
	// - If there are non-boolean parts (e.g. "User") in the union, they are always truthy
	//   In that case, true branch becomes nonBool ∪ {"true"} and false branch becomes {"false"}
	// - If only the boolean part is present, then narrow to {"true"} and {"false"} respectively
	if currentType.Contains("bool") || currentType.Contains("true") || currentType.Contains("false") {
		nonBool := currentType.Clone().Erase("bool").Erase("true").Erase("false")
		if nonBool.Len() > 0 {
			if currentType.Contains("bool") || currentType.Contains("true") {
				trueType = nonBool.Union(types.NewMap("true"))
			} else {
				trueType = nonBool
			}
			falseType = types.NewMap("false")
		} else {
			trueType = types.NewMap("true")
			falseType = types.NewMap("false")
		}
	}

	// Note: For other types (e.g. int, string, array), our type system doesn't include literal values,
	// so we don't perform additional narrowing

	// If we are in the "not" context (i.e. if(!$variable)), swap the branches
	if a.inNot {
		trueType, falseType = falseType, trueType
	}

	a.trueContext.sc.ReplaceVar(variable, trueType, "type narrowing for "+variable.Name, meta.VarAlwaysDefined)
	a.falseContext.sc.ReplaceVar(variable, falseType, "type narrowing for "+variable.Name, meta.VarAlwaysDefined)
}

func (a *andWalker) handleTypeCheckCondition(expectedType string, args []ir.Node) {
	for _, arg := range args {
		argument, ok := arg.(*ir.Argument)
		if !ok {
			continue
		}
		variable, ok := argument.Expr.(*ir.SimpleVar)
		if !ok {
			continue
		}

		// Traverse the variable to ensure it exists, since this variable
		// will be added to the context later
		a.b.handleVariable(variable)

		// Get the current type of the variable from the appropriate context
		currentType := a.exprType(variable) // nolint:staticcheck
		if a.inNot {
			currentType = a.exprTypeInContext(a.trueContext, variable)
		} else {
			currentType = a.exprTypeInContext(a.falseContext, variable)
		}

		var trueType, falseType types.Map

		switch expectedType {
		case "bool":
			// For bool: consider possible literal types "bool", "true" and "false"
			boolMerge := types.MergeMaps(types.NewMap("bool"), types.NewMap("true"), types.NewMap("false"))
			intersection := currentType.Intersect(boolMerge)
			if intersection.Empty() {
				// If there is no explicit bool subtype, then the positive branch becomes simply "bool"
				trueType = types.NewMap("bool")
			} else {
				// Otherwise, keep exactly those literals that were present in the current type
				trueType = intersection
			}
			// Negative branch: remove all bool subtypes
			falseType = currentType.Clone().Erase("bool").Erase("true").Erase("false")
		case "object":
			// For is_object: keep only keys that are not considered primitive
			keys := currentType.Keys()
			var objectKeys []string
			for _, k := range keys {
				switch k {
				case "int", "float", "string", "bool", "null", "true", "false", "mixed", "callable", "resource", "void", "iterable", "never":
					// Skip primitive types
					continue
				default:
					objectKeys = append(objectKeys, k)
				}
			}
			if len(objectKeys) == 0 {
				trueType = types.NewMap("object")
			} else {
				trueType = types.NewEmptyMap(1)
				for _, k := range objectKeys {
					trueType = trueType.Union(types.NewMap(k))
				}
			}
			falseType = currentType.Clone()
			for _, k := range objectKeys {
				falseType = falseType.Erase(k)
			}
		default:
			// Standard logic for other types
			trueType = types.NewMap(expectedType)
			falseType = currentType.Clone().Erase(expectedType)
		}

		if a.inNot {
			trueType, falseType = falseType, trueType
		}

		a.trueContext.sc.ReplaceVar(variable, trueType, "type narrowing for "+expectedType, meta.VarAlwaysDefined)
		a.falseContext.sc.ReplaceVar(variable, falseType, "type narrowing for "+expectedType, meta.VarAlwaysDefined)
	}
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

	currentVar, isGotVar := a.b.ctx.sc.GetVar(variable)
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
