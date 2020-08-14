package normalize

import (
	"github.com/VKCOM/noverify/src/ir"
)

func sideEffectFree(n ir.Node) bool {
	f := sideEffectsFinder{}
	n.Walk(&f)
	return !f.sideEffects
}

type sideEffectsFinder struct {
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

	return false
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
		f.sideEffects = true
		return false

	case *ir.StaticCallExpr:
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
