package solver

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
)

func SideEffectFreeFunc(sc *meta.Scope, st *meta.ClassParseState, customTypes []CustomType, stmts []ir.Node) bool {
	// TODO: functions that call pure functions are also pure.
	// TODO: allow local var assignments those RHS is pure.
	f := sideEffectsFinder{sc: sc, st: st, customTypes: customTypes}
	n := &ir.StmtList{Stmts: stmts}
	n.Walk(&f)
	return !f.sideEffects
}

// SideEffectFree reports whether n is a side effect free expression.
//
// If indexing is completed, some function calls may be permitted as well
// if they're proven to be side effect free as well.
func SideEffectFree(sc *meta.Scope, st *meta.ClassParseState, customTypes []CustomType, n ir.Node) bool {
	f := sideEffectsFinder{sc: sc, st: st, customTypes: customTypes}
	n.Walk(&f)
	return !f.sideEffects
}

type sideEffectsFinder struct {
	sc          *meta.Scope
	st          *meta.ClassParseState
	customTypes []CustomType

	sideEffects bool
}

var pureBuiltins = func() map[string]struct{} {
	list := []string{
		`\count`,
		`\sizeof`,

		// Array functions.
		`\array_key_exists`,
		`\array_keys`,
		`\array_merge`,
		`\array_slice`,
		`\array_values`,
		`\explode`,
		`\implode`,
		`\in_array`,

		// String functions.
		`\str_replace`,
		`\strlen`,
		`\strpos`,
		`\strtolower`,
		`\strtoupper`,
		`\substr`,
		`\trim`,

		// Simple math functions.
		`\abs`,
		`\floor`,
		`\max`,
		`\min`,

		// All type converting functions are pure.
		`\boolval`,
		`\doubleval`,
		`\floatval`,
		`\intval`,
		`\strval`,

		// All type predicates are pure.
		`\is_array`,
		`\is_bool`,
		`\is_callable`,
		`\is_countable`,
		`\is_double`,
		`\is_float`,
		`\is_int`,
		`\is_integer`,
		`\is_iterable`,
		`\is_long`,
		`\is_null`,
		`\is_numeric`,
		`\is_object`,
		`\is_real`,
		`\is_resource`,
		`\is_scalar`,
		`\is_string`,
	}

	set := make(map[string]struct{}, len(list))
	for _, nm := range list {
		set[nm] = struct{}{}
	}
	return set
}()

func (f *sideEffectsFinder) functionCallIsPure(n *ir.FunctionCallExpr) bool {
	// We allow referencing builtin funcs even before indexing is completed.

	var funcName string
	nm, ok := n.Function.(*ir.Name)
	if ok {
		if nm.IsFullyQualified() {
			funcName = nm.Value
		} else {
			funcName = `\` + nm.Value
		}
	}

	if funcName != "" {
		// Might be a builtin.
		if _, ok := pureBuiltins[funcName]; ok {
			return true
		}
	}

	if !meta.IsIndexingComplete() {
		return false
	}
	if funcName != "" {
		// We can't properly annotate builtin funcs
		// as pure during the indexing, since we don't have
		// their PHP sources.
		_, ok := meta.GetInternalFunctionInfo(funcName)
		if ok {
			return false
		}
	}

	fqName, ok := GetFuncName(f.st, n.Function)
	if !ok {
		return false
	}
	info, ok := meta.Info.GetFunction(fqName)
	if !ok {
		return false
	}
	return info.IsPure() && info.ExitFlags == 0
}

func (f *sideEffectsFinder) staticCallIsPure(n *ir.StaticCallExpr) bool {
	if !meta.IsIndexingComplete() {
		return false
	}
	methodName, ok := n.Call.(*ir.Identifier)
	if !ok {
		return false
	}
	className, ok := GetClassName(f.st, n.Class)
	if !ok {
		return false
	}
	m, ok := FindMethod(className, methodName.Value)
	return ok && m.Info.IsPure() && m.Info.ExitFlags == 0
}

func (f *sideEffectsFinder) methodCallIsPure(n *ir.MethodCallExpr) bool {
	if !meta.IsIndexingComplete() {
		return false
	}

	methodName, ok := n.Method.(*ir.Identifier)
	if !ok {
		return false
	}

	types := ExprTypeCustom(f.sc, f.st, n.Variable, f.customTypes)
	if types.Len() != 1 || types.Is("mixed") {
		return false
	}

	return types.Find(func(typ meta.Type) bool {
		m, ok := FindMethod(typ.String(), methodName.Value)
		if meta.IsInternalClass(m.ClassName) {
			return false
		}

		return ok && m.Info.IsPure() && m.Info.ExitFlags == 0
	})
}

func (f *sideEffectsFinder) EnterNode(n ir.Node) bool {
	if f.sideEffects {
		return false
	}

	// We can get false positives for overloaded operations.
	// For example, array index can be an offsetGet() call,
	// which might not be pure.

	switch n := n.(type) {
	case *ir.FunctionCallExpr:
		if f.functionCallIsPure(n) {
			return true
		}
		f.sideEffects = true
		return false

	case *ir.MethodCallExpr:
		if f.methodCallIsPure(n) {
			return true
		}
		f.sideEffects = true
		return false

	case *ir.StaticCallExpr:
		if f.staticCallIsPure(n) {
			return true
		}
		f.sideEffects = true
		return false

	case *ir.PrintExpr,
		*ir.EchoStmt,
		*ir.UnsetStmt,
		*ir.ThrowStmt,
		*ir.GlobalStmt,
		*ir.ExitExpr,
		*ir.Assign,
		*ir.AssignReference,
		*ir.AssignBitwiseAnd,
		*ir.AssignBitwiseOr,
		*ir.AssignBitwiseXor,
		*ir.AssignConcat,
		*ir.AssignDiv,
		*ir.AssignMinus,
		*ir.AssignMod,
		*ir.AssignMul,
		*ir.AssignPlus,
		*ir.AssignPow,
		*ir.AssignShiftLeft,
		*ir.AssignShiftRight,
		*ir.YieldExpr,
		*ir.YieldFromExpr,
		*ir.EvalExpr,
		*ir.PreIncExpr,
		*ir.PostIncExpr,
		*ir.PreDecExpr,
		*ir.PostDecExpr,
		*ir.ImportExpr:
		f.sideEffects = true
		return false
	}

	return true
}

func (f *sideEffectsFinder) LeaveNode(n ir.Node) {}
