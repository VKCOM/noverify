package constfold

import (
	"strconv"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

// Eval tries to compute the e using the constant expressions folding.
// In case of failure, meta.UnknownValue is returned.
func Eval(st *meta.ClassParseState, e ir.Node) meta.ConstantValue {
	// TODO: support more operators and some builtin PHP functions like strlen.

	switch e := e.(type) {
	case *ir.ParenExpr:
		return Eval(st, e.Expr)

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

	case *ir.PlusExpr:
		return Plus(Eval(st, e.Left), Eval(st, e.Right))
	case *ir.MulExpr:
		return Mul(Eval(st, e.Left), Eval(st, e.Right))

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
		v := irutil.Unquote(e.Value)
		return meta.ConstantValue{Value: v, Type: meta.String}
	}

	return meta.UnknownValue
}
