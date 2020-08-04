package solver

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

func SideEffectFreeFunc(sc *meta.Scope, st *meta.ClassParseState, customTypes []CustomType, stmts []node.Node) bool {
	// TODO: functions that call pure functions are also pure.
	// TODO: allow local var assignments those RHS is pure.
	f := sideEffectsFinder{sc: sc, st: st, customTypes: customTypes}
	n := &stmt.StmtList{Stmts: stmts}
	n.Walk(&f)
	return !f.sideEffects
}

// SideEffectFree reports whether n is a side effect free expression.
//
// If indexing is completed, some function calls may be permitted as well
// if they're proven to be side effect free as well.
func SideEffectFree(sc *meta.Scope, st *meta.ClassParseState, customTypes []CustomType, n node.Node) bool {
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

func (f *sideEffectsFinder) functionCallIsPure(n *expr.FunctionCall) bool {
	// We allow referencing builtin funcs even before indexing is completed.

	var funcName string
	switch nm := n.Function.(type) {
	case *name.Name:
		if len(nm.Parts) == 1 {
			funcName = `\` + meta.NameToString(nm)
		}
	case *name.FullyQualified:
		if len(nm.Parts) == 1 {
			funcName = meta.FullyQualifiedToString(nm)
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

func (f *sideEffectsFinder) staticCallIsPure(n *expr.StaticCall) bool {
	if !meta.IsIndexingComplete() {
		return false
	}
	methodName, ok := n.Call.(*node.Identifier)
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

func (f *sideEffectsFinder) methodCallIsPure(n *expr.MethodCall) bool {
	if !meta.IsIndexingComplete() {
		return false
	}
	methodName, ok := n.Method.(*node.Identifier)
	if !ok {
		return false
	}
	typ := ExprTypeCustom(f.sc, f.st, n.Variable, f.customTypes)
	if typ.Len() != 1 || typ.Is("mixed") {
		return false
	}
	return typ.Find(func(typ string) bool {
		m, ok := FindMethod(typ, methodName.Value)
		if meta.IsInternalClass(m.ClassName) {
			return false
		}
		return ok && m.Info.IsPure() && m.Info.ExitFlags == 0
	})
}

func (f *sideEffectsFinder) EnterNode(w walker.Walkable) bool {
	if f.sideEffects {
		return false
	}

	// We can get false positives for overloaded operations.
	// For example, array index can be an offsetGet() call,
	// which might not be pure.

	switch n := w.(type) {
	case *expr.FunctionCall:
		if f.functionCallIsPure(n) {
			return true
		}
		f.sideEffects = true
		return false

	case *expr.MethodCall:
		if f.methodCallIsPure(n) {
			return true
		}
		f.sideEffects = true
		return false

	case *expr.StaticCall:
		if f.staticCallIsPure(n) {
			return true
		}
		f.sideEffects = true
		return false

	case *expr.Print,
		*stmt.Echo,
		*stmt.Unset,
		*stmt.Throw,
		*stmt.Global,
		*expr.Exit,
		*assign.Assign,
		*assign.Reference,
		*assign.BitwiseAnd,
		*assign.BitwiseOr,
		*assign.BitwiseXor,
		*assign.Concat,
		*assign.Div,
		*assign.Minus,
		*assign.Mod,
		*assign.Mul,
		*assign.Plus,
		*assign.Pow,
		*assign.ShiftLeft,
		*assign.ShiftRight,
		*expr.Yield,
		*expr.YieldFrom,
		*expr.Eval,
		*expr.PreInc,
		*expr.PostInc,
		*expr.PreDec,
		*expr.PostDec,
		*expr.Require,
		*expr.RequireOnce,
		*expr.Include,
		*expr.IncludeOnce:
		f.sideEffects = true
		return false
	}

	return true
}

func (f *sideEffectsFinder) LeaveNode(w walker.Walkable) {}
