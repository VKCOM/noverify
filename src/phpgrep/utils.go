package phpgrep

import (
	"github.com/z7zmey/php-parser/pkg/position"

	"github.com/VKCOM/noverify/src/ir"
)

func findNamed(capture []CapturedNode, name string) (ir.Node, bool) {
	for _, c := range capture {
		if c.Name == name {
			return c.Node, true
		}
	}
	return nil, false
}

func getNodePos(n ir.Node) *position.Position {
	pos := ir.GetPosition(n)
	if pos == nil {
		// FIXME: investigate how and why we're getting nil position for some nodes.
		// See #24.
		return nil
	}
	if pos.EndPos < 0 || pos.StartPos < 0 {
		// FIXME: investigate why we sometimes get out-of-range pos ranges.
		// We also get negative EndPos for some nodes, which is awkward.
		// See #24.
		return nil
	}
	return pos
}

func nodeIsVar(n ir.Node) bool {
	switch n.(type) {
	case *ir.SimpleVar, *ir.Var:
		return true
	default:
		return false
	}
}

func nodeIsExpr(n ir.Node) bool {
	switch n.(type) {
	case *ir.Assign,
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
		*ir.AssignReference,
		*ir.AssignShiftLeft,
		*ir.AssignShiftRight,
		*ir.Var,
		*ir.SimpleVar,
		*ir.BitwiseAndExpr,
		*ir.BitwiseOrExpr,
		*ir.BitwiseXorExpr,
		*ir.BooleanAndExpr,
		*ir.BooleanOrExpr,
		*ir.CoalesceExpr,
		*ir.ConcatExpr,
		*ir.DivExpr,
		*ir.EqualExpr,
		*ir.GreaterExpr,
		*ir.GreaterOrEqualExpr,
		*ir.IdenticalExpr,
		*ir.LogicalAndExpr,
		*ir.LogicalOrExpr,
		*ir.LogicalXorExpr,
		*ir.MinusExpr,
		*ir.ModExpr,
		*ir.MulExpr,
		*ir.NotEqualExpr,
		*ir.NotIdenticalExpr,
		*ir.PlusExpr,
		*ir.PowExpr,
		*ir.ShiftLeftExpr,
		*ir.ShiftRightExpr,
		*ir.SmallerExpr,
		*ir.SmallerOrEqualExpr,
		*ir.SpaceshipExpr,
		*ir.ArrayExpr,
		*ir.ArrayDimFetchExpr,
		*ir.ArrayItemExpr,
		*ir.BitwiseNotExpr,
		*ir.BooleanNotExpr,
		*ir.ClassConstFetchExpr,
		*ir.CloneExpr,
		*ir.ClosureExpr,
		*ir.ClosureUsesExpr,
		*ir.ConstFetchExpr,
		*ir.EmptyExpr,
		*ir.ErrorSuppressExpr,
		*ir.EvalExpr,
		*ir.ExitExpr,
		*ir.FunctionCallExpr,
		*ir.ImportExpr,
		*ir.InstanceOfExpr,
		*ir.IssetExpr,
		*ir.MethodCallExpr,
		*ir.NewExpr,
		*ir.PostDecExpr,
		*ir.PreIncExpr,
		*ir.PrintExpr,
		*ir.PropertyFetchExpr,
		*ir.ReferenceExpr,
		*ir.ShellExecExpr,
		*ir.StaticCallExpr,
		*ir.StaticPropertyFetchExpr,
		*ir.TernaryExpr,
		*ir.UnaryMinusExpr,
		*ir.UnaryPlusExpr,
		*ir.YieldExpr,
		*ir.YieldFromExpr,
		*ir.Dnumber,
		*ir.Encapsed,
		*ir.EncapsedStringPart,
		*ir.Heredoc,
		*ir.Lnumber,
		*ir.MagicConstant,
		*ir.String,
		*ir.ExpressionStmt:
		return true

	default:
		return false
	}
}

func matchMetaVar(n ir.Node, s string) bool {
	switch n := n.(type) {
	case *ir.ArrayItemExpr:
		return n.Key == nil && matchMetaVar(n.Val, s)
	case *ir.ExpressionStmt:
		return matchMetaVar(n.Expr, s)
	case *ir.Argument:
		return matchMetaVar(n.Expr, s)

	case *ir.Var:
		nm, ok := n.Expr.(*ir.String)
		return ok && nm.Value == s

	default:
		return false
	}
}
