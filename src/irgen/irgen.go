package irgen

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/cast"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
)

func ConvertRoot(n *node.Root) *ir.Root {
	return ConvertNode(n).(*ir.Root)
}

func ConvertNodeSlice(xs []node.Node) []ir.Node {
	out := make([]ir.Node, len(xs))
	for i, x := range xs {
		out[i] = ConvertNode(x)
	}
	return out

}

func ConvertNode(n node.Node) ir.Node {
	if n == nil {
		return nil
	}
	switch n := n.(type) {
	case *assign.Assign:
		if n == nil {
			return (*ir.Assign)(nil)
		}
		out := &ir.Assign{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.BitwiseAnd:
		if n == nil {
			return (*ir.AssignBitwiseAnd)(nil)
		}
		out := &ir.AssignBitwiseAnd{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.BitwiseOr:
		if n == nil {
			return (*ir.AssignBitwiseOr)(nil)
		}
		out := &ir.AssignBitwiseOr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.BitwiseXor:
		if n == nil {
			return (*ir.AssignBitwiseXor)(nil)
		}
		out := &ir.AssignBitwiseXor{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Coalesce:
		if n == nil {
			return (*ir.AssignCoalesce)(nil)
		}
		out := &ir.AssignCoalesce{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Concat:
		if n == nil {
			return (*ir.AssignConcat)(nil)
		}
		out := &ir.AssignConcat{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Div:
		if n == nil {
			return (*ir.AssignDiv)(nil)
		}
		out := &ir.AssignDiv{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Minus:
		if n == nil {
			return (*ir.AssignMinus)(nil)
		}
		out := &ir.AssignMinus{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Mod:
		if n == nil {
			return (*ir.AssignMod)(nil)
		}
		out := &ir.AssignMod{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Mul:
		if n == nil {
			return (*ir.AssignMul)(nil)
		}
		out := &ir.AssignMul{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Plus:
		if n == nil {
			return (*ir.AssignPlus)(nil)
		}
		out := &ir.AssignPlus{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Pow:
		if n == nil {
			return (*ir.AssignPow)(nil)
		}
		out := &ir.AssignPow{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.Reference:
		if n == nil {
			return (*ir.AssignReference)(nil)
		}
		out := &ir.AssignReference{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.ShiftLeft:
		if n == nil {
			return (*ir.AssignShiftLeft)(nil)
		}
		out := &ir.AssignShiftLeft{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *assign.ShiftRight:
		if n == nil {
			return (*ir.AssignShiftRight)(nil)
		}
		out := &ir.AssignShiftRight{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Expression = ConvertNode(n.Expression)
		return out

	case *binary.BitwiseAnd:
		if n == nil {
			return (*ir.BitwiseAndExpr)(nil)
		}
		out := &ir.BitwiseAndExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.BitwiseOr:
		if n == nil {
			return (*ir.BitwiseOrExpr)(nil)
		}
		out := &ir.BitwiseOrExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.BitwiseXor:
		if n == nil {
			return (*ir.BitwiseXorExpr)(nil)
		}
		out := &ir.BitwiseXorExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.BooleanAnd:
		if n == nil {
			return (*ir.BooleanAndExpr)(nil)
		}
		out := &ir.BooleanAndExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.BooleanOr:
		if n == nil {
			return (*ir.BooleanOrExpr)(nil)
		}
		out := &ir.BooleanOrExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Coalesce:
		if n == nil {
			return (*ir.CoalesceExpr)(nil)
		}
		out := &ir.CoalesceExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Concat:
		if n == nil {
			return (*ir.ConcatExpr)(nil)
		}
		out := &ir.ConcatExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Div:
		if n == nil {
			return (*ir.DivExpr)(nil)
		}
		out := &ir.DivExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Equal:
		if n == nil {
			return (*ir.EqualExpr)(nil)
		}
		out := &ir.EqualExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Greater:
		if n == nil {
			return (*ir.GreaterExpr)(nil)
		}
		out := &ir.GreaterExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.GreaterOrEqual:
		if n == nil {
			return (*ir.GreaterOrEqualExpr)(nil)
		}
		out := &ir.GreaterOrEqualExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Identical:
		if n == nil {
			return (*ir.IdenticalExpr)(nil)
		}
		out := &ir.IdenticalExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.LogicalAnd:
		if n == nil {
			return (*ir.LogicalAndExpr)(nil)
		}
		out := &ir.LogicalAndExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.LogicalOr:
		if n == nil {
			return (*ir.LogicalOrExpr)(nil)
		}
		out := &ir.LogicalOrExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.LogicalXor:
		if n == nil {
			return (*ir.LogicalXorExpr)(nil)
		}
		out := &ir.LogicalXorExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Minus:
		if n == nil {
			return (*ir.MinusExpr)(nil)
		}
		out := &ir.MinusExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Mod:
		if n == nil {
			return (*ir.ModExpr)(nil)
		}
		out := &ir.ModExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Mul:
		if n == nil {
			return (*ir.MulExpr)(nil)
		}
		out := &ir.MulExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.NotEqual:
		if n == nil {
			return (*ir.NotEqualExpr)(nil)
		}
		out := &ir.NotEqualExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.NotIdentical:
		if n == nil {
			return (*ir.NotIdenticalExpr)(nil)
		}
		out := &ir.NotIdenticalExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Plus:
		if n == nil {
			return (*ir.PlusExpr)(nil)
		}
		out := &ir.PlusExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Pow:
		if n == nil {
			return (*ir.PowExpr)(nil)
		}
		out := &ir.PowExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.ShiftLeft:
		if n == nil {
			return (*ir.ShiftLeftExpr)(nil)
		}
		out := &ir.ShiftLeftExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.ShiftRight:
		if n == nil {
			return (*ir.ShiftRightExpr)(nil)
		}
		out := &ir.ShiftRightExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Smaller:
		if n == nil {
			return (*ir.SmallerExpr)(nil)
		}
		out := &ir.SmallerExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.SmallerOrEqual:
		if n == nil {
			return (*ir.SmallerOrEqualExpr)(nil)
		}
		out := &ir.SmallerOrEqualExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *binary.Spaceship:
		if n == nil {
			return (*ir.SpaceshipExpr)(nil)
		}
		out := &ir.SpaceshipExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Left = ConvertNode(n.Left)
		out.Right = ConvertNode(n.Right)
		return out

	case *cast.Array:
		return convCastExpr(n, n.Expr, "array")
	case *cast.Bool:
		return convCastExpr(n, n.Expr, "bool")
	case *cast.Int:
		return convCastExpr(n, n.Expr, "int")
	case *cast.Double:
		return convCastExpr(n, n.Expr, "float")
	case *cast.Object:
		return convCastExpr(n, n.Expr, "object")
	case *cast.String:
		return convCastExpr(n, n.Expr, "string")

	case *cast.Unset:
		// We dont convert (unset)$x into CastExpr deliberately.
		if n == nil {
			return (*ir.UnsetCastExpr)(nil)
		}
		out := &ir.UnsetCastExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.Array:
		if n == nil {
			return (*ir.ArrayExpr)(nil)
		}
		out := &ir.ArrayExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		{
			slice := make([]*ir.ArrayItemExpr, len(n.Items))
			for i := range n.Items {
				slice[i] = ConvertNode(n.Items[i]).(*ir.ArrayItemExpr)
			}
			out.Items = slice
		}
		out.ShortSyntax = n.ShortSyntax
		return out

	case *expr.ArrayDimFetch:
		if n == nil {
			return (*ir.ArrayDimFetchExpr)(nil)
		}
		out := &ir.ArrayDimFetchExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Dim = ConvertNode(n.Dim)
		return out

	case *expr.ArrayItem:
		if n == nil {
			return (*ir.ArrayItemExpr)(nil)
		}
		out := &ir.ArrayItemExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Key = ConvertNode(n.Key)
		out.Val = ConvertNode(n.Val)
		out.Unpack = n.Unpack
		return out

	case *expr.ArrowFunction:
		if n == nil {
			return (*ir.ArrowFunctionExpr)(nil)
		}
		out := &ir.ArrowFunctionExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.ReturnsRef = n.ReturnsRef
		out.Static = n.Static
		out.PhpDocComment = n.PhpDocComment
		out.Params = ConvertNodeSlice(n.Params)
		out.ReturnType = ConvertNode(n.ReturnType)
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.BitwiseNot:
		if n == nil {
			return (*ir.BitwiseNotExpr)(nil)
		}
		out := &ir.BitwiseNotExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.BooleanNot:
		if n == nil {
			return (*ir.BooleanNotExpr)(nil)
		}
		out := &ir.BooleanNotExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.ClassConstFetch:
		if n == nil {
			return (*ir.ClassConstFetchExpr)(nil)
		}
		out := &ir.ClassConstFetchExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Class = ConvertNode(n.Class)
		out.ConstantName = ConvertNode(n.ConstantName).(*ir.Identifier)
		return out

	case *expr.Clone:
		if n == nil {
			return (*ir.CloneExpr)(nil)
		}
		out := &ir.CloneExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.Closure:
		if n == nil {
			return (*ir.ClosureExpr)(nil)
		}
		out := &ir.ClosureExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.ReturnsRef = n.ReturnsRef
		out.Static = n.Static
		out.PhpDocComment = n.PhpDocComment
		out.Params = ConvertNodeSlice(n.Params)
		out.ClosureUse = ConvertNode(n.ClosureUse).(*ir.ClosureUseExpr)
		out.ReturnType = ConvertNode(n.ReturnType)
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *expr.ClosureUse:
		if n == nil {
			return (*ir.ClosureUseExpr)(nil)
		}
		out := &ir.ClosureUseExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Uses = ConvertNodeSlice(n.Uses)
		return out

	case *expr.ConstFetch:
		if n == nil {
			return (*ir.ConstFetchExpr)(nil)
		}
		out := &ir.ConstFetchExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Constant = ConvertNode(n.Constant)
		return out

	case *expr.Empty:
		if n == nil {
			return (*ir.EmptyExpr)(nil)
		}
		out := &ir.EmptyExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.ErrorSuppress:
		if n == nil {
			return (*ir.ErrorSuppressExpr)(nil)
		}
		out := &ir.ErrorSuppressExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.Eval:
		if n == nil {
			return (*ir.EvalExpr)(nil)
		}
		out := &ir.EvalExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.Exit:
		if n == nil {
			return (*ir.ExitExpr)(nil)
		}
		out := &ir.ExitExpr{}
		out.FreeFloating = n.FreeFloating
		out.Die = n.Die
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.FunctionCall:
		if n == nil {
			return (*ir.FunctionCallExpr)(nil)
		}
		out := &ir.FunctionCallExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Function = ConvertNode(n.Function)
		out.ArgumentList = ConvertNode(n.ArgumentList).(*ir.ArgumentList)
		return out

	case *expr.InstanceOf:
		if n == nil {
			return (*ir.InstanceOfExpr)(nil)
		}
		out := &ir.InstanceOfExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		out.Class = ConvertNode(n.Class)
		return out

	case *expr.Isset:
		if n == nil {
			return (*ir.IssetExpr)(nil)
		}
		out := &ir.IssetExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variables = ConvertNodeSlice(n.Variables)
		return out

	case *expr.List:
		if n == nil {
			return (*ir.ListExpr)(nil)
		}
		out := &ir.ListExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		{
			slice := make([]*ir.ArrayItemExpr, len(n.Items))
			for i := range n.Items {
				slice[i] = ConvertNode(n.Items[i]).(*ir.ArrayItemExpr)
			}
			out.Items = slice
		}
		out.ShortSyntax = n.ShortSyntax
		return out

	case *expr.MethodCall:
		if n == nil {
			return (*ir.MethodCallExpr)(nil)
		}
		out := &ir.MethodCallExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Method = ConvertNode(n.Method)
		out.ArgumentList = ConvertNode(n.ArgumentList).(*ir.ArgumentList)
		return out

	case *expr.New:
		if n == nil {
			return (*ir.NewExpr)(nil)
		}
		out := &ir.NewExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Class = ConvertNode(n.Class)
		out.ArgumentList = ConvertNode(n.ArgumentList).(*ir.ArgumentList)
		return out

	case *expr.Paren:
		if n == nil {
			return (*ir.ParenExpr)(nil)
		}
		out := &ir.ParenExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.PostDec:
		if n == nil {
			return (*ir.PostDecExpr)(nil)
		}
		out := &ir.PostDecExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		return out

	case *expr.PostInc:
		if n == nil {
			return (*ir.PostIncExpr)(nil)
		}
		out := &ir.PostIncExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		return out

	case *expr.PreDec:
		if n == nil {
			return (*ir.PreDecExpr)(nil)
		}
		out := &ir.PreDecExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		return out

	case *expr.PreInc:
		if n == nil {
			return (*ir.PreIncExpr)(nil)
		}
		out := &ir.PreIncExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		return out

	case *expr.Print:
		if n == nil {
			return (*ir.PrintExpr)(nil)
		}
		out := &ir.PrintExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.PropertyFetch:
		if n == nil {
			return (*ir.PropertyFetchExpr)(nil)
		}
		out := &ir.PropertyFetchExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		out.Property = ConvertNode(n.Property)
		return out

	case *expr.Reference:
		if n == nil {
			return (*ir.ReferenceExpr)(nil)
		}
		out := &ir.ReferenceExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable)
		return out

	case *expr.Require:
		return convImportExpr(n, n.Expr, "require")
	case *expr.RequireOnce:
		return convImportExpr(n, n.Expr, "require_once")
	case *expr.Include:
		return convImportExpr(n, n.Expr, "include")
	case *expr.IncludeOnce:
		return convImportExpr(n, n.Expr, "include_once")

	case *expr.ShellExec:
		if n == nil {
			return (*ir.ShellExecExpr)(nil)
		}
		out := &ir.ShellExecExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Parts = ConvertNodeSlice(n.Parts)
		return out

	case *expr.StaticCall:
		if n == nil {
			return (*ir.StaticCallExpr)(nil)
		}
		out := &ir.StaticCallExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Class = ConvertNode(n.Class)
		out.Call = ConvertNode(n.Call)
		out.ArgumentList = ConvertNode(n.ArgumentList).(*ir.ArgumentList)
		return out

	case *expr.StaticPropertyFetch:
		if n == nil {
			return (*ir.StaticPropertyFetchExpr)(nil)
		}
		out := &ir.StaticPropertyFetchExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Class = ConvertNode(n.Class)
		out.Property = ConvertNode(n.Property)
		return out

	case *expr.Ternary:
		if n == nil {
			return (*ir.TernaryExpr)(nil)
		}
		out := &ir.TernaryExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Condition = ConvertNode(n.Condition)
		out.IfTrue = ConvertNode(n.IfTrue)
		out.IfFalse = ConvertNode(n.IfFalse)
		return out

	case *expr.UnaryMinus:
		if n == nil {
			return (*ir.UnaryMinusExpr)(nil)
		}
		out := &ir.UnaryMinusExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.UnaryPlus:
		if n == nil {
			return (*ir.UnaryPlusExpr)(nil)
		}
		out := &ir.UnaryPlusExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *expr.Yield:
		if n == nil {
			return (*ir.YieldExpr)(nil)
		}
		out := &ir.YieldExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Key = ConvertNode(n.Key)
		out.Value = ConvertNode(n.Value)
		return out

	case *expr.YieldFrom:
		if n == nil {
			return (*ir.YieldFromExpr)(nil)
		}
		out := &ir.YieldFromExpr{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *name.FullyQualified:
		if n == nil {
			return (*ir.FullyQualifiedName)(nil)
		}
		out := &ir.FullyQualifiedName{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Parts = ConvertNodeSlice(n.Parts)
		return out

	case *name.Name:
		if n == nil {
			return (*ir.Name)(nil)
		}
		out := &ir.Name{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Parts = ConvertNodeSlice(n.Parts)
		return out

	case *name.NamePart:
		if n == nil {
			return (*ir.NamePart)(nil)
		}
		out := &ir.NamePart{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Value = n.Value
		return out

	case *name.Relative:
		if n == nil {
			return (*ir.RelativeName)(nil)
		}
		out := &ir.RelativeName{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Parts = ConvertNodeSlice(n.Parts)
		return out

	case *node.Argument:
		if n == nil {
			return (*ir.Argument)(nil)
		}
		out := &ir.Argument{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variadic = n.Variadic
		out.IsReference = n.IsReference
		out.Expr = ConvertNode(n.Expr)
		return out

	case *node.ArgumentList:
		if n == nil {
			return (*ir.ArgumentList)(nil)
		}
		out := &ir.ArgumentList{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		{
			slice := make([]ir.Node, len(n.Arguments))
			for i := range n.Arguments {
				slice[i] = ConvertNode(n.Arguments[i])
			}
			out.Arguments = slice
		}
		return out

	case *node.Identifier:
		if n == nil {
			return (*ir.Identifier)(nil)
		}
		out := &ir.Identifier{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Value = n.Value
		return out

	case *node.Nullable:
		if n == nil {
			return (*ir.Nullable)(nil)
		}
		out := &ir.Nullable{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *node.Parameter:
		if n == nil {
			return (*ir.Parameter)(nil)
		}
		out := &ir.Parameter{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.ByRef = n.ByRef
		out.Variadic = n.Variadic
		out.VariableType = ConvertNode(n.VariableType)
		out.Variable = ConvertNode(n.Variable).(*ir.SimpleVar)
		out.DefaultValue = ConvertNode(n.DefaultValue)
		return out

	case *node.Root:
		if n == nil {
			return (*ir.Root)(nil)
		}
		out := &ir.Root{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		{
			slice := make([]ir.Node, len(n.Stmts))
			for i := range n.Stmts {
				slice[i] = ConvertNode(n.Stmts[i])
			}
			out.Stmts = slice
		}
		return out

	case *node.SimpleVar:
		if n == nil {
			return (*ir.SimpleVar)(nil)
		}
		out := &ir.SimpleVar{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Name = n.Name
		return out

	case *node.Var:
		if n == nil {
			return (*ir.Var)(nil)
		}
		out := &ir.Var{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *scalar.Dnumber:
		if n == nil {
			return (*ir.Dnumber)(nil)
		}
		out := &ir.Dnumber{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Value = n.Value
		return out

	case *scalar.Encapsed:
		if n == nil {
			return (*ir.Encapsed)(nil)
		}
		out := &ir.Encapsed{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Parts = ConvertNodeSlice(n.Parts)
		return out

	case *scalar.EncapsedStringPart:
		if n == nil {
			return (*ir.EncapsedStringPart)(nil)
		}
		out := &ir.EncapsedStringPart{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Value = n.Value
		return out

	case *scalar.Heredoc:
		if n == nil {
			return (*ir.Heredoc)(nil)
		}
		out := &ir.Heredoc{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Label = n.Label
		out.Parts = ConvertNodeSlice(n.Parts)
		return out

	case *scalar.Lnumber:
		if n == nil {
			return (*ir.Lnumber)(nil)
		}
		out := &ir.Lnumber{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Value = n.Value
		return out

	case *scalar.MagicConstant:
		if n == nil {
			return (*ir.MagicConstant)(nil)
		}
		out := &ir.MagicConstant{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Value = n.Value
		return out

	case *scalar.String:
		if n == nil {
			return (*ir.String)(nil)
		}
		out := &ir.String{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Value = n.Value
		return out

	case *stmt.Break:
		if n == nil {
			return (*ir.BreakStmt)(nil)
		}
		out := &ir.BreakStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *stmt.Case:
		if n == nil {
			return (*ir.CaseStmt)(nil)
		}
		out := &ir.CaseStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Cond = ConvertNode(n.Cond)
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.CaseList:
		if n == nil {
			return (*ir.CaseListStmt)(nil)
		}
		out := &ir.CaseListStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Cases = ConvertNodeSlice(n.Cases)
		return out

	case *stmt.Catch:
		if n == nil {
			return (*ir.CatchStmt)(nil)
		}
		out := &ir.CatchStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Types = ConvertNodeSlice(n.Types)
		out.Variable = ConvertNode(n.Variable).(*ir.SimpleVar)
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.Class:
		if n == nil {
			return (*ir.ClassStmt)(nil)
		}
		out := &ir.ClassStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.PhpDocComment = n.PhpDocComment
		out.ClassName = ConvertNode(n.ClassName).(*ir.Identifier)
		{
			slice := make([]*ir.Identifier, len(n.Modifiers))
			for i := range n.Modifiers {
				slice[i] = ConvertNode(n.Modifiers[i]).(*ir.Identifier)
			}
			out.Modifiers = slice
		}
		out.ArgumentList = ConvertNode(n.ArgumentList).(*ir.ArgumentList)
		out.Extends = ConvertNode(n.Extends).(*ir.ClassExtendsStmt)
		out.Implements = ConvertNode(n.Implements).(*ir.ClassImplementsStmt)
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.ClassConstList:
		if n == nil {
			return (*ir.ClassConstListStmt)(nil)
		}
		out := &ir.ClassConstListStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		{
			slice := make([]*ir.Identifier, len(n.Modifiers))
			for i := range n.Modifiers {
				slice[i] = ConvertNode(n.Modifiers[i]).(*ir.Identifier)
			}
			out.Modifiers = slice
		}
		out.Consts = ConvertNodeSlice(n.Consts)
		return out

	case *stmt.ClassExtends:
		if n == nil {
			return (*ir.ClassExtendsStmt)(nil)
		}
		out := &ir.ClassExtendsStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.ClassName = ConvertNode(n.ClassName)
		return out

	case *stmt.ClassImplements:
		if n == nil {
			return (*ir.ClassImplementsStmt)(nil)
		}
		out := &ir.ClassImplementsStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.InterfaceNames = ConvertNodeSlice(n.InterfaceNames)
		return out

	case *stmt.ClassMethod:
		if n == nil {
			return (*ir.ClassMethodStmt)(nil)
		}
		out := &ir.ClassMethodStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.ReturnsRef = n.ReturnsRef
		out.PhpDocComment = n.PhpDocComment
		out.MethodName = ConvertNode(n.MethodName).(*ir.Identifier)
		{
			slice := make([]*ir.Identifier, len(n.Modifiers))
			for i := range n.Modifiers {
				slice[i] = ConvertNode(n.Modifiers[i]).(*ir.Identifier)
			}
			out.Modifiers = slice
		}
		out.Params = ConvertNodeSlice(n.Params)
		out.ReturnType = ConvertNode(n.ReturnType)
		out.Stmt = ConvertNode(n.Stmt)
		return out

	case *stmt.ConstList:
		if n == nil {
			return (*ir.ConstListStmt)(nil)
		}
		out := &ir.ConstListStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Consts = ConvertNodeSlice(n.Consts)
		return out

	case *stmt.Constant:
		if n == nil {
			return (*ir.ConstantStmt)(nil)
		}
		out := &ir.ConstantStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.PhpDocComment = n.PhpDocComment
		out.ConstantName = ConvertNode(n.ConstantName).(*ir.Identifier)
		out.Expr = ConvertNode(n.Expr)
		return out

	case *stmt.Continue:
		if n == nil {
			return (*ir.ContinueStmt)(nil)
		}
		out := &ir.ContinueStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *stmt.Declare:
		if n == nil {
			return (*ir.DeclareStmt)(nil)
		}
		out := &ir.DeclareStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Consts = ConvertNodeSlice(n.Consts)
		out.Stmt = ConvertNode(n.Stmt)
		out.Alt = n.Alt
		return out

	case *stmt.Default:
		if n == nil {
			return (*ir.DefaultStmt)(nil)
		}
		out := &ir.DefaultStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.Do:
		if n == nil {
			return (*ir.DoStmt)(nil)
		}
		out := &ir.DoStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Stmt = ConvertNode(n.Stmt)
		out.Cond = ConvertNode(n.Cond)
		return out

	case *stmt.Echo:
		if n == nil {
			return (*ir.EchoStmt)(nil)
		}
		out := &ir.EchoStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Exprs = ConvertNodeSlice(n.Exprs)
		return out

	case *stmt.Else:
		if n == nil {
			return (*ir.ElseStmt)(nil)
		}
		out := &ir.ElseStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Stmt = ConvertNode(n.Stmt)
		out.AltSyntax = n.AltSyntax
		return out

	case *stmt.ElseIf:
		if n == nil {
			return (*ir.ElseIfStmt)(nil)
		}
		out := &ir.ElseIfStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Cond = ConvertNode(n.Cond)
		out.Stmt = ConvertNode(n.Stmt)
		out.AltSyntax = n.AltSyntax
		out.Merged = n.Merged
		return out

	case *stmt.Expression:
		if n == nil {
			return (*ir.ExpressionStmt)(nil)
		}
		out := &ir.ExpressionStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *stmt.Finally:
		if n == nil {
			return (*ir.FinallyStmt)(nil)
		}
		out := &ir.FinallyStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.For:
		if n == nil {
			return (*ir.ForStmt)(nil)
		}
		out := &ir.ForStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Init = ConvertNodeSlice(n.Init)
		out.Cond = ConvertNodeSlice(n.Cond)
		out.Loop = ConvertNodeSlice(n.Loop)
		out.Stmt = ConvertNode(n.Stmt)
		out.AltSyntax = n.AltSyntax
		return out

	case *stmt.Foreach:
		if n == nil {
			return (*ir.ForeachStmt)(nil)
		}
		out := &ir.ForeachStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		out.Key = ConvertNode(n.Key)
		out.Variable = ConvertNode(n.Variable)
		out.Stmt = ConvertNode(n.Stmt)
		out.AltSyntax = n.AltSyntax
		return out

	case *stmt.Function:
		if n == nil {
			return (*ir.FunctionStmt)(nil)
		}
		out := &ir.FunctionStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.ReturnsRef = n.ReturnsRef
		out.PhpDocComment = n.PhpDocComment
		out.FunctionName = ConvertNode(n.FunctionName).(*ir.Identifier)
		out.Params = ConvertNodeSlice(n.Params)
		out.ReturnType = ConvertNode(n.ReturnType)
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.Global:
		if n == nil {
			return (*ir.GlobalStmt)(nil)
		}
		out := &ir.GlobalStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Vars = ConvertNodeSlice(n.Vars)
		return out

	case *stmt.Goto:
		if n == nil {
			return (*ir.GotoStmt)(nil)
		}
		out := &ir.GotoStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Label = ConvertNode(n.Label).(*ir.Identifier)
		return out

	case *stmt.GroupUse:
		if n == nil {
			return (*ir.GroupUseStmt)(nil)
		}
		out := &ir.GroupUseStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.UseType = ConvertNode(n.UseType)
		out.Prefix = ConvertNode(n.Prefix)
		out.UseList = ConvertNodeSlice(n.UseList)
		return out

	case *stmt.HaltCompiler:
		if n == nil {
			return (*ir.HaltCompilerStmt)(nil)
		}
		out := &ir.HaltCompilerStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		return out

	case *stmt.If:
		if n == nil {
			return (*ir.IfStmt)(nil)
		}
		out := &ir.IfStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Cond = ConvertNode(n.Cond)
		out.Stmt = ConvertNode(n.Stmt)
		out.ElseIf = ConvertNodeSlice(n.ElseIf)
		out.Else = ConvertNode(n.Else)
		out.AltSyntax = n.AltSyntax
		return out

	case *stmt.InlineHtml:
		if n == nil {
			return (*ir.InlineHTMLStmt)(nil)
		}
		out := &ir.InlineHTMLStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Value = n.Value
		return out

	case *stmt.Interface:
		if n == nil {
			return (*ir.InterfaceStmt)(nil)
		}
		out := &ir.InterfaceStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.PhpDocComment = n.PhpDocComment
		out.InterfaceName = ConvertNode(n.InterfaceName).(*ir.Identifier)
		out.Extends = ConvertNode(n.Extends).(*ir.InterfaceExtendsStmt)
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.InterfaceExtends:
		if n == nil {
			return (*ir.InterfaceExtendsStmt)(nil)
		}
		out := &ir.InterfaceExtendsStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.InterfaceNames = ConvertNodeSlice(n.InterfaceNames)
		return out

	case *stmt.Label:
		if n == nil {
			return (*ir.LabelStmt)(nil)
		}
		out := &ir.LabelStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.LabelName = ConvertNode(n.LabelName).(*ir.Identifier)
		return out

	case *stmt.Namespace:
		if n == nil {
			return (*ir.NamespaceStmt)(nil)
		}
		out := &ir.NamespaceStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.NamespaceName = ConvertNode(n.NamespaceName)
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.Nop:
		if n == nil {
			return (*ir.NopStmt)(nil)
		}
		out := &ir.NopStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		return out

	case *stmt.Property:
		if n == nil {
			return (*ir.PropertyStmt)(nil)
		}
		out := &ir.PropertyStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.PhpDocComment = n.PhpDocComment
		out.Variable = ConvertNode(n.Variable).(*ir.SimpleVar)
		out.Expr = ConvertNode(n.Expr)
		return out

	case *stmt.PropertyList:
		if n == nil {
			return (*ir.PropertyListStmt)(nil)
		}
		out := &ir.PropertyListStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		{
			slice := make([]*ir.Identifier, len(n.Modifiers))
			for i := range n.Modifiers {
				slice[i] = ConvertNode(n.Modifiers[i]).(*ir.Identifier)
			}
			out.Modifiers = slice
		}
		out.Type = ConvertNode(n.Type)
		out.Properties = ConvertNodeSlice(n.Properties)
		return out

	case *stmt.Return:
		if n == nil {
			return (*ir.ReturnStmt)(nil)
		}
		out := &ir.ReturnStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *stmt.Static:
		if n == nil {
			return (*ir.StaticStmt)(nil)
		}
		out := &ir.StaticStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Vars = ConvertNodeSlice(n.Vars)
		return out

	case *stmt.StaticVar:
		if n == nil {
			return (*ir.StaticVarStmt)(nil)
		}
		out := &ir.StaticVarStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Variable = ConvertNode(n.Variable).(*ir.SimpleVar)
		out.Expr = ConvertNode(n.Expr)
		return out

	case *stmt.StmtList:
		if n == nil {
			return (*ir.StmtList)(nil)
		}
		out := &ir.StmtList{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.Switch:
		if n == nil {
			return (*ir.SwitchStmt)(nil)
		}
		out := &ir.SwitchStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Cond = ConvertNode(n.Cond)
		out.CaseList = ConvertNode(n.CaseList).(*ir.CaseListStmt)
		out.AltSyntax = n.AltSyntax
		return out

	case *stmt.Throw:
		if n == nil {
			return (*ir.ThrowStmt)(nil)
		}
		out := &ir.ThrowStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Expr = ConvertNode(n.Expr)
		return out

	case *stmt.Trait:
		if n == nil {
			return (*ir.TraitStmt)(nil)
		}
		out := &ir.TraitStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.PhpDocComment = n.PhpDocComment
		out.TraitName = ConvertNode(n.TraitName).(*ir.Identifier)
		out.Stmts = ConvertNodeSlice(n.Stmts)
		return out

	case *stmt.TraitAdaptationList:
		if n == nil {
			return (*ir.TraitAdaptationListStmt)(nil)
		}
		out := &ir.TraitAdaptationListStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Adaptations = ConvertNodeSlice(n.Adaptations)
		return out

	case *stmt.TraitMethodRef:
		if n == nil {
			return (*ir.TraitMethodRefStmt)(nil)
		}
		out := &ir.TraitMethodRefStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Trait = ConvertNode(n.Trait)
		out.Method = ConvertNode(n.Method).(*ir.Identifier)
		return out

	case *stmt.TraitUse:
		if n == nil {
			return (*ir.TraitUseStmt)(nil)
		}
		out := &ir.TraitUseStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Traits = ConvertNodeSlice(n.Traits)
		out.TraitAdaptationList = ConvertNode(n.TraitAdaptationList)
		return out

	case *stmt.TraitUseAlias:
		if n == nil {
			return (*ir.TraitUseAliasStmt)(nil)
		}
		out := &ir.TraitUseAliasStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Ref = ConvertNode(n.Ref)
		out.Modifier = ConvertNode(n.Modifier)
		out.Alias = ConvertNode(n.Alias).(*ir.Identifier)
		return out

	case *stmt.TraitUsePrecedence:
		if n == nil {
			return (*ir.TraitUsePrecedenceStmt)(nil)
		}
		out := &ir.TraitUsePrecedenceStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Ref = ConvertNode(n.Ref)
		out.Insteadof = ConvertNodeSlice(n.Insteadof)
		return out

	case *stmt.Try:
		if n == nil {
			return (*ir.TryStmt)(nil)
		}
		out := &ir.TryStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Stmts = ConvertNodeSlice(n.Stmts)
		out.Catches = ConvertNodeSlice(n.Catches)
		out.Finally = ConvertNode(n.Finally)
		return out

	case *stmt.Unset:
		if n == nil {
			return (*ir.UnsetStmt)(nil)
		}
		out := &ir.UnsetStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Vars = ConvertNodeSlice(n.Vars)
		return out

	case *stmt.Use:
		if n == nil {
			return (*ir.UseStmt)(nil)
		}
		out := &ir.UseStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.UseType = ConvertNode(n.UseType).(*ir.Identifier)
		out.Use = ConvertNode(n.Use)
		out.Alias = ConvertNode(n.Alias).(*ir.Identifier)
		return out

	case *stmt.UseList:
		if n == nil {
			return (*ir.UseListStmt)(nil)
		}
		out := &ir.UseListStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.UseType = ConvertNode(n.UseType)
		out.Uses = ConvertNodeSlice(n.Uses)
		return out

	case *stmt.While:
		if n == nil {
			return (*ir.WhileStmt)(nil)
		}
		out := &ir.WhileStmt{}
		out.FreeFloating = n.FreeFloating
		out.Position = n.Position
		out.Cond = ConvertNode(n.Cond)
		out.Stmt = ConvertNode(n.Stmt)
		out.AltSyntax = n.AltSyntax
		return out
	}

	panic(fmt.Sprintf("unhandled type %T", n))
}

func convImportExpr(n, e node.Node, fn string) *ir.ImportExpr {
	return &ir.ImportExpr{
		FreeFloating: *n.GetFreeFloating(),
		Position:     n.GetPosition(),
		Func:         fn,
		Expr:         ConvertNode(e),
	}
}

func convCastExpr(n, e node.Node, typ string) *ir.TypeCastExpr {
	return &ir.TypeCastExpr{
		FreeFloating: *n.GetFreeFloating(),
		Position:     n.GetPosition(),
		Type:         typ,
		Expr:         ConvertNode(e),
	}
}
