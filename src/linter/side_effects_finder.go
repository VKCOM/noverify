package linter

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/solver"
)

func sideEffectFreeFunc(sc *meta.Scope, st *meta.ClassParseState, customTypes []solver.CustomType, stmts []node.Node) bool {
	// TODO: functions that call pure functions are also pure.
	// TODO: allow local var assignments those RHS is pure.
	f := sideEffectsFinder{sc: sc, st: st, customTypes: customTypes}
	n := &stmt.StmtList{Stmts: stmts}
	n.Walk(&f)
	return !f.sideEffects
}

func sideEffectFree(sc *meta.Scope, st *meta.ClassParseState, customTypes []solver.CustomType, n node.Node) bool {
	f := sideEffectsFinder{sc: sc, st: st, customTypes: customTypes}
	n.Walk(&f)
	return !f.sideEffects
}

type sideEffectsFinder struct {
	sc          *meta.Scope
	st          *meta.ClassParseState
	customTypes []solver.CustomType

	sideEffects bool
}

var pureBuiltins = func() map[string]struct{} {
	list := []string{
		`\count`,
		`\strlen`,
		`\implode`,
		`\explode`,
	}

	set := make(map[string]struct{}, len(list))
	for _, nm := range list {
		set[nm] = struct{}{}
	}
	return set
}()

func (f *sideEffectsFinder) functionCallIsPure(n *expr.FunctionCall) bool {
	// Allow referencing builtin funcs even before indexing is completed.
	nm, ok := n.Function.(*name.Name)
	var funcName string
	if ok && len(nm.Parts) == 1 {
		// Might be a builtin.
		funcName = `\` + meta.NameToString(nm)
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

	call := resolveFunctionCall(f.sc, f.st, f.customTypes, n)
	if !call.defined || !call.canAnalyze || call.fqName == "" {
		return false
	}
	return call.info.IsPure() && call.info.ExitFlags == 0
}

func (f *sideEffectsFinder) staticCallIsPure(n *expr.StaticCall) bool {
	if !meta.IsIndexingComplete() {
		return false
	}
	methodName, ok := n.Call.(*node.Identifier)
	if !ok {
		return false
	}
	className, ok := solver.GetClassName(f.st, n.Class)
	if !ok {
		return false
	}
	m, ok := solver.FindMethod(className, methodName.Value)
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
	typ := solver.ExprTypeCustom(f.sc, f.st, n.Variable, f.customTypes)
	if typ.Len() != 1 || typ.Is("mixed") {
		return false
	}
	return typ.Find(func(typ string) bool {
		m, ok := solver.FindMethod(typ, methodName.Value)
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
		// fmt.Printf("node=%T %s\n", w, FmtNode(w.(node.Node)))
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
