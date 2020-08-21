package constfold

import (
	"strconv"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

// Eval tries to compute the e using the constant expressions folding.
// In case of failure, meta.UnknownValue is returned.
func Eval(st *meta.ClassParseState, e ir.Node) meta.ConstantValue {
	// TODO: support more operators and some builtin PHP functions like strlen.

	switch e := e.(type) {
	case *ir.Argument:
		return Eval(st, e.Expr)
	case *ir.ParenExpr:
		return Eval(st, e.Expr)

	case *ir.ClassConstFetchExpr:
		if !meta.IsIndexingComplete() {
			return meta.UnknownValue
		}
		className, ok := solver.GetClassName(st, e.Class)
		if !ok {
			return meta.UnknownValue
		}
		info, _, ok := solver.FindConstant(className, e.ConstantName.Value)
		if !ok {
			return meta.UnknownValue
		}
		return info.Value

	case *ir.ConstFetchExpr:
		switch e.Constant.Value {
		case `true`:
			return meta.TrueValue
		case `false`:
			return meta.FalseValue
		}
		if !meta.IsIndexingComplete() {
			return meta.UnknownValue
		}
		_, info, ok := solver.GetConstant(st, e.Constant)
		if !ok {
			return meta.UnknownValue
		}
		return info.Value

	case *ir.UnaryMinusExpr:
		return Neg(Eval(st, e.Expr))

	case *ir.BooleanNotExpr:
		return Not(Eval(st, e.Expr))

	case *ir.BooleanAndExpr:
		return And(Eval(st, e.Left), Eval(st, e.Right))
	case *ir.BooleanOrExpr:
		return Or(Eval(st, e.Left), Eval(st, e.Right))
	case *ir.LogicalAndExpr:
		return And(Eval(st, e.Left), Eval(st, e.Right))
	case *ir.LogicalOrExpr:
		return Or(Eval(st, e.Left), Eval(st, e.Right))

	case *ir.PlusExpr:
		return Plus(Eval(st, e.Left), Eval(st, e.Right))
	case *ir.MinusExpr:
		return Minus(Eval(st, e.Left), Eval(st, e.Right))
	case *ir.MulExpr:
		return Mul(Eval(st, e.Left), Eval(st, e.Right))
	case *ir.ConcatExpr:
		return Concat(Eval(st, e.Left), Eval(st, e.Right))

	case *ir.Lnumber:
		value, err := strconv.ParseInt(e.Value, 10, 64)
		if err != nil {
			return meta.UnknownValue
		}
		return meta.ConstantValue{Type: meta.Integer, Value: value}

	case *ir.Dnumber:
		value, err := strconv.ParseFloat(e.Value, 64)
		if err != nil {
			return meta.UnknownValue
		}
		return meta.ConstantValue{Type: meta.Float, Value: value}

	case *ir.String:
		return meta.ConstantValue{Value: e.Value, Type: meta.String}
	}

	return meta.UnknownValue
}
