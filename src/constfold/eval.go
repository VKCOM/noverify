package constfold

import (
	"path/filepath"
	"strconv"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/solver"
)

// Eval tries to compute the e using the constant expressions folding.
// In case of failure, meta.UnknownValue and false flag is returned.
func Eval(st *meta.ClassParseState, e ir.Node) meta.ConstValue {
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

	case *ir.BitwiseAndExpr:
		return BitAnd(Eval(st, e.Left), Eval(st, e.Right))
	case *ir.BitwiseOrExpr:
		return BitOr(Eval(st, e.Left), Eval(st, e.Right))

	case *ir.Lnumber:
		value, err := strconv.ParseInt(e.Value, 0, 64)
		if err != nil {
			return meta.UnknownValue
		}
		return meta.NewIntConst(value)

	case *ir.Dnumber:
		value, err := strconv.ParseFloat(e.Value, 64)
		if err != nil {
			return meta.UnknownValue
		}
		return meta.NewFloatConst(value)

	case *ir.String:
		return meta.NewStringConst(e.Value)

	case *ir.FunctionCallExpr:
		// dirname(__FILE__)
		if !meta.NameNodeEquals(e.Function, `dirname`) {
			return meta.UnknownValue
		}
		if len(e.Args) == 0 {
			return meta.UnknownValue
		}
		arg, ok := e.Arg(0).Expr.(*ir.MagicConstant)
		if !ok || arg.Value != "__FILE__" {
			return meta.UnknownValue
		}
		return meta.NewStringConst(filepath.Dir(st.CurrentFile))

	case *ir.MagicConstant:
		switch e.Value {
		case "__LINE__":
			return meta.NewIntConst(int64(e.Position.StartLine))
		case "__FILE__":
			return meta.NewStringConst(st.CurrentFile)
		case "__DIR__":
			return meta.NewStringConst(filepath.Dir(st.CurrentFile))
		case "__FUNCTION__":
			return meta.NewStringConst(st.CurrentFunction)
		case "__METHOD__":
			return meta.NewStringConst(st.CurrentClass + "::" + st.CurrentFunction)
		case "__CLASS__":
			return meta.NewStringConst(st.CurrentClass)
		case "__NAMESPACE__":
			return meta.NewStringConst(st.Namespace)
		case "__TRAIT__":
			if st.IsTrait {
				return meta.NewStringConst(st.CurrentClass)
			}
		}
	}

	return meta.UnknownValue
}
