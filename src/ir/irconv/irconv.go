package irconv

import (
	"bytes"
	"fmt"

	"github.com/z7zmey/php-parser/pkg/ast"
	"github.com/z7zmey/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
	"github.com/VKCOM/noverify/src/phpdoc"
)

func ConvertNode(n ast.Vertex) ir.Node {
	c := NewConverter(phpdoc.NewTypeParser())
	return c.ConvertNode(n)
}

type Converter struct {
	namespace string

	phpdocTypeParser *phpdoc.TypeParser
}

// NewConverter returns a new AST->IR converter.
//
// If typeParser is nil, it will not eagerly try to parse phpdoc
// strings into phpdoc.CommentPart.
//
// It's intended to be re-used inside a signle thread context.
func NewConverter(typeParser *phpdoc.TypeParser) *Converter {
	return &Converter{
		phpdocTypeParser: typeParser,
	}
}

func (c *Converter) ConvertRoot(n *ast.Root) *ir.Root {
	return c.ConvertNode(n).(*ir.Root)
}

func (c *Converter) ConvertNode(n ast.Vertex) ir.Node {
	c.reset()
	return c.convNode(n)
}

func (c *Converter) reset() {
	c.namespace = ""
}

func (c *Converter) convNodeSlice(xs []ast.Vertex) []ir.Node {
	out := make([]ir.Node, len(xs))
	for i, x := range xs {
		out[i] = c.convNode(x)
	}
	return out
}

func (c *Converter) convNode(n ast.Vertex) ir.Node {
	if n == nil {
		return nil
	}
	switch n := n.(type) {
	case *ast.ExprAssign:
		if n == nil {
			return (*ir.Assign)(nil)
		}
		out := &ir.Assign{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)

		// hack for expressions like:
		// /**
		//  * @param Boo $x
		// */
		// $_ = fn($x) => $x->b();
		if arrowFn, ok := out.Expression.(*ir.ArrowFunctionExpr); ok {
			doc, found := irutil.FindPhpDoc(out.Variable)

			if found {
				arrowFn.PhpDocComment = doc
				arrowFn.PhpDoc = c.parsePHPDoc(doc)
			}
		}

		return out

	case *ast.ExprAssignBitwiseAnd:
		if n == nil {
			return (*ir.AssignBitwiseAnd)(nil)
		}
		out := &ir.AssignBitwiseAnd{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignBitwiseOr:
		if n == nil {
			return (*ir.AssignBitwiseOr)(nil)
		}
		out := &ir.AssignBitwiseOr{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignBitwiseXor:
		if n == nil {
			return (*ir.AssignBitwiseXor)(nil)
		}
		out := &ir.AssignBitwiseXor{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignCoalesce:
		if n == nil {
			return (*ir.AssignCoalesce)(nil)
		}
		out := &ir.AssignCoalesce{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignConcat:
		if n == nil {
			return (*ir.AssignConcat)(nil)
		}
		out := &ir.AssignConcat{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignDiv:
		if n == nil {
			return (*ir.AssignDiv)(nil)
		}
		out := &ir.AssignDiv{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignMinus:
		if n == nil {
			return (*ir.AssignMinus)(nil)
		}
		out := &ir.AssignMinus{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignMod:
		if n == nil {
			return (*ir.AssignMod)(nil)
		}
		out := &ir.AssignMod{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignMul:
		if n == nil {
			return (*ir.AssignMul)(nil)
		}
		out := &ir.AssignMul{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignPlus:
		if n == nil {
			return (*ir.AssignPlus)(nil)
		}
		out := &ir.AssignPlus{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignPow:
		if n == nil {
			return (*ir.AssignPow)(nil)
		}
		out := &ir.AssignPow{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignReference:
		if n == nil {
			return (*ir.AssignReference)(nil)
		}
		out := &ir.AssignReference{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignShiftLeft:
		if n == nil {
			return (*ir.AssignShiftLeft)(nil)
		}
		out := &ir.AssignShiftLeft{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprAssignShiftRight:
		if n == nil {
			return (*ir.AssignShiftRight)(nil)
		}
		out := &ir.AssignShiftRight{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var)
		out.Expression = c.convNode(n.Expr)
		return out

	case *ast.ExprBinaryBitwiseAnd:
		if n == nil {
			return (*ir.BitwiseAndExpr)(nil)
		}
		out := &ir.BitwiseAndExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryBitwiseOr:
		if n == nil {
			return (*ir.BitwiseOrExpr)(nil)
		}
		out := &ir.BitwiseOrExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryBitwiseXor:
		if n == nil {
			return (*ir.BitwiseXorExpr)(nil)
		}
		out := &ir.BitwiseXorExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryBooleanAnd:
		if n == nil {
			return (*ir.BooleanAndExpr)(nil)
		}
		out := &ir.BooleanAndExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryBooleanOr:
		if n == nil {
			return (*ir.BooleanOrExpr)(nil)
		}
		out := &ir.BooleanOrExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryCoalesce:
		if n == nil {
			return (*ir.CoalesceExpr)(nil)
		}
		out := &ir.CoalesceExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryConcat:
		if n == nil {
			return (*ir.ConcatExpr)(nil)
		}
		out := &ir.ConcatExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryDiv:
		if n == nil {
			return (*ir.DivExpr)(nil)
		}
		out := &ir.DivExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryEqual:
		if n == nil {
			return (*ir.EqualExpr)(nil)
		}
		out := &ir.EqualExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryGreater:
		if n == nil {
			return (*ir.GreaterExpr)(nil)
		}
		out := &ir.GreaterExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryGreaterOrEqual:
		if n == nil {
			return (*ir.GreaterOrEqualExpr)(nil)
		}
		out := &ir.GreaterOrEqualExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryIdentical:
		if n == nil {
			return (*ir.IdenticalExpr)(nil)
		}
		out := &ir.IdenticalExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryLogicalAnd:
		if n == nil {
			return (*ir.LogicalAndExpr)(nil)
		}
		out := &ir.LogicalAndExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryLogicalOr:
		if n == nil {
			return (*ir.LogicalOrExpr)(nil)
		}
		out := &ir.LogicalOrExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryLogicalXor:
		if n == nil {
			return (*ir.LogicalXorExpr)(nil)
		}
		out := &ir.LogicalXorExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryMinus:
		if n == nil {
			return (*ir.MinusExpr)(nil)
		}
		out := &ir.MinusExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryMod:
		if n == nil {
			return (*ir.ModExpr)(nil)
		}
		out := &ir.ModExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryMul:
		if n == nil {
			return (*ir.MulExpr)(nil)
		}
		out := &ir.MulExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryNotEqual:
		if n == nil {
			return (*ir.NotEqualExpr)(nil)
		}
		out := &ir.NotEqualExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryNotIdentical:
		if n == nil {
			return (*ir.NotIdenticalExpr)(nil)
		}
		out := &ir.NotIdenticalExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryPlus:
		if n == nil {
			return (*ir.PlusExpr)(nil)
		}
		out := &ir.PlusExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryPow:
		if n == nil {
			return (*ir.PowExpr)(nil)
		}
		out := &ir.PowExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryShiftLeft:
		if n == nil {
			return (*ir.ShiftLeftExpr)(nil)
		}
		out := &ir.ShiftLeftExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinaryShiftRight:
		if n == nil {
			return (*ir.ShiftRightExpr)(nil)
		}
		out := &ir.ShiftRightExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinarySmaller:
		if n == nil {
			return (*ir.SmallerExpr)(nil)
		}
		out := &ir.SmallerExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinarySmallerOrEqual:
		if n == nil {
			return (*ir.SmallerOrEqualExpr)(nil)
		}
		out := &ir.SmallerOrEqualExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprBinarySpaceship:
		if n == nil {
			return (*ir.SpaceshipExpr)(nil)
		}
		out := &ir.SpaceshipExpr{}
		out.Position = n.Position
		out.OpTkn = n.OpTkn
		out.Left = c.convNode(n.Left)
		out.Right = c.convNode(n.Right)
		return out

	case *ast.ExprCastArray:
		return c.convCastExpr(n, n.Expr, n.CastTkn, "array")
	case *ast.ExprCastBool:
		return c.convCastExpr(n, n.Expr, n.CastTkn, "bool")
	case *ast.ExprCastInt:
		return c.convCastExpr(n, n.Expr, n.CastTkn, "int")
	case *ast.ExprCastDouble:
		return c.convCastExpr(n, n.Expr, n.CastTkn, "float")
	case *ast.ExprCastObject:
		return c.convCastExpr(n, n.Expr, n.CastTkn, "object")
	case *ast.ExprCastString:
		return c.convCastExpr(n, n.Expr, n.CastTkn, "string")

	case *ast.ExprCastUnset:
		// We dont convert (unset)$x into CastExpr deliberately.
		if n == nil {
			return (*ir.UnsetCastExpr)(nil)
		}
		out := &ir.UnsetCastExpr{}
		out.Position = n.Position
		out.CastTkn = n.CastTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprArray:
		if n == nil {
			return (*ir.ArrayExpr)(nil)
		}
		out := &ir.ArrayExpr{}
		out.Position = n.Position
		out.ArrayTkn = n.ArrayTkn
		out.OpenBracketTkn = n.OpenBracketTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseBracketTkn = n.CloseBracketTkn
		{
			slice := make([]*ir.ArrayItemExpr, len(n.Items))
			for i := range n.Items {
				slice[i] = c.convNode(n.Items[i]).(*ir.ArrayItemExpr)
			}
			out.Items = slice
		}
		out.ShortSyntax = !hasValue(n.ArrayTkn)
		return out

	case *ast.ExprArrayDimFetch:
		if n == nil {
			return (*ir.ArrayDimFetchExpr)(nil)
		}
		out := &ir.ArrayDimFetchExpr{}
		out.Position = n.Position
		out.OpenBracketTkn = n.OpenBracketTkn
		out.CloseBracketTkn = n.CloseBracketTkn
		out.Variable = c.convNode(n.Var)
		out.Dim = c.convNode(n.Dim)

		out.CurlyBrace = hasValue(out.OpenBracketTkn) && out.OpenBracketTkn.Value[0] == '{'

		return out

	case *ast.ExprArrayItem:
		if n == nil {
			return (*ir.ArrayItemExpr)(nil)
		}
		out := &ir.ArrayItemExpr{}
		out.Position = n.Position
		out.EllipsisTkn = n.EllipsisTkn
		out.DoubleArrowTkn = n.DoubleArrowTkn
		out.AmpersandTkn = n.AmpersandTkn

		out.Key = c.convNode(n.Key)

		if hasValue(n.AmpersandTkn) {
			out.Val = &ir.ReferenceExpr{
				FreeFloating: nil,
				AmpersandTkn: n.AmpersandTkn,
				Position:     n.Position,
				Variable:     c.convNode(n.Val),
			}
		} else {
			out.Val = c.convNode(n.Val)
		}

		out.Unpack = hasValue(n.EllipsisTkn)
		return out

	case *ast.ExprArrowFunction:
		if n == nil {
			return (*ir.ArrowFunctionExpr)(nil)
		}
		out := &ir.ArrowFunctionExpr{}
		out.Position = n.Position

		out.StaticTkn = n.StaticTkn
		out.FnTkn = n.FnTkn
		out.AmpersandTkn = n.AmpersandTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn

		out.ReturnsRef = hasValue(n.AmpersandTkn)
		out.Static = hasValue(n.StaticTkn)

		var tokenWithDoc *token.Token
		if n.StaticTkn != nil {
			tokenWithDoc = n.StaticTkn
		} else {
			tokenWithDoc = n.FnTkn
		}

		out.PhpDocComment, out.PhpDoc = c.getPhpDocWithParse(tokenWithDoc)

		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn

		out.Params = c.convNodeSlice(n.Params)
		out.ReturnType = c.convNode(n.ReturnType)

		out.DoubleArrowTkn = n.DoubleArrowTkn

		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprBitwiseNot:
		if n == nil {
			return (*ir.BitwiseNotExpr)(nil)
		}
		out := &ir.BitwiseNotExpr{}
		out.Position = n.Position
		out.TildaTkn = n.TildaTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprBooleanNot:
		if n == nil {
			return (*ir.BooleanNotExpr)(nil)
		}
		out := &ir.BooleanNotExpr{}
		out.Position = n.Position
		out.ExclamationTkn = n.ExclamationTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprClassConstFetch:
		if n == nil {
			return (*ir.ClassConstFetchExpr)(nil)
		}
		out := &ir.ClassConstFetchExpr{}
		out.Position = n.Position
		out.DoubleColonTkn = n.DoubleColonTkn
		out.Class = c.convNode(n.Class)
		out.ConstantName = c.convNode(n.Const).(*ir.Identifier)
		return out

	case *ast.ExprClone:
		if n == nil {
			return (*ir.CloneExpr)(nil)
		}
		out := &ir.CloneExpr{}
		out.Position = n.Position
		out.CloneTkn = n.CloneTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprClosure:
		if n == nil {
			return (*ir.ClosureExpr)(nil)
		}
		out := &ir.ClosureExpr{}
		out.Position = n.Position

		out.StaticTkn = n.StaticTkn
		out.FunctionTkn = n.FunctionTkn
		out.AmpersandTkn = n.AmpersandTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.UseTkn = n.UseTkn
		out.UseOpenParenthesisTkn = n.UseOpenParenthesisTkn
		out.UseSeparatorTkns = n.UseSeparatorTkns
		out.UseCloseParenthesisTkn = n.UseCloseParenthesisTkn
		out.ColonTkn = n.ColonTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn

		var tokenWithDoc *token.Token
		if n.StaticTkn != nil {
			tokenWithDoc = n.StaticTkn
		} else {
			tokenWithDoc = n.FunctionTkn
		}

		out.PhpDocComment, out.PhpDoc = c.getPhpDocWithParse(tokenWithDoc)

		out.ReturnsRef = hasValue(n.AmpersandTkn)
		out.Static = hasValue(n.StaticTkn)

		out.Params = c.convNodeSlice(n.Params)
		out.ClosureUse = &ir.ClosureUseExpr{
			FreeFloating: nil,
			Position:     nil,
			Uses:         c.convNodeSlice(n.Uses),
		}
		out.ReturnType = c.convNode(n.ReturnType)
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.ExprClosureUse:
		if n == nil {
			return (*ir.SimpleVar)(nil)
		}

		varNode := c.convNode(n.Var)

		if hasValue(n.AmpersandTkn) {
			varNode = &ir.ReferenceExpr{
				FreeFloating: nil,
				AmpersandTkn: n.AmpersandTkn,
				Position:     n.Position,
				Variable:     varNode,
			}
		}

		switch varNode := varNode.(type) {
		case *ir.SimpleVar:
			varNode.Position = n.Position
		case *ir.Var:
			varNode.Position = n.Position
		}

		return varNode

	case *ast.ExprConstFetch:
		if n == nil {
			return (*ir.ConstFetchExpr)(nil)
		}
		out := &ir.ConstFetchExpr{}
		out.Position = n.Position
		out.Constant = c.convNode(n.Const).(*ir.Name)
		return out

	case *ast.ExprEmpty:
		if n == nil {
			return (*ir.EmptyExpr)(nil)
		}
		out := &ir.EmptyExpr{}
		out.Position = n.Position
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprErrorSuppress:
		if n == nil {
			return (*ir.ErrorSuppressExpr)(nil)
		}
		out := &ir.ErrorSuppressExpr{}
		out.Position = n.Position
		out.AtTkn = n.AtTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprEval:
		if n == nil {
			return (*ir.EvalExpr)(nil)
		}
		out := &ir.EvalExpr{}
		out.Position = n.Position
		out.EvalTkn = n.EvalTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprExit:
		if n == nil {
			return (*ir.ExitExpr)(nil)
		}
		out := &ir.ExitExpr{}
		out.Position = n.Position
		out.ExitTkn = n.ExitTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.Expr = c.convNode(n.Expr)

		out.Die = hasValue(n.ExitTkn) && bytes.Equal(n.ExitTkn.Value, []byte("die"))
		return out

	case *ast.ExprFunctionCall:
		if n == nil {
			return (*ir.FunctionCallExpr)(nil)
		}
		out := &ir.FunctionCallExpr{}
		out.Position = n.Position
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.Function = c.convNode(n.Function)
		out.Args = c.convNodeSlice(n.Args)
		return out

	case *ast.ExprInstanceOf:
		if n == nil {
			return (*ir.InstanceOfExpr)(nil)
		}
		out := &ir.InstanceOfExpr{}
		out.Position = n.Position
		out.InstanceOfTkn = n.InstanceOfTkn
		out.Expr = c.convNode(n.Expr)
		out.Class = c.convNode(n.Class)
		return out

	case *ast.ExprIsset:
		if n == nil {
			return (*ir.IssetExpr)(nil)
		}
		out := &ir.IssetExpr{}
		out.Position = n.Position
		out.IssetTkn = n.IssetTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.Variables = c.convNodeSlice(n.Vars)
		return out

	case *ast.ExprList:
		if n == nil {
			return (*ir.ListExpr)(nil)
		}
		out := &ir.ListExpr{}
		out.Position = n.Position

		out.ListTkn = n.ListTkn
		out.OpenBracketTkn = n.OpenBracketTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseBracketTkn = n.CloseBracketTkn
		{
			slice := make([]*ir.ArrayItemExpr, len(n.Items))
			for i := range n.Items {
				slice[i] = c.convNode(n.Items[i]).(*ir.ArrayItemExpr)
			}
			out.Items = slice
		}
		out.ShortSyntax = !hasValue(n.ListTkn)
		return out

	case *ast.ExprMethodCall:
		if n == nil {
			return (*ir.MethodCallExpr)(nil)
		}
		out := &ir.MethodCallExpr{}
		out.Position = n.Position
		out.ObjectOperatorTkn = n.ObjectOperatorTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.Variable = c.convNode(n.Var)
		out.Method = c.convNode(n.Method)
		out.Args = c.convNodeSlice(n.Args)
		return out

	case *ast.ExprNew:
		if n == nil {
			return (*ir.NewExpr)(nil)
		}
		out := &ir.NewExpr{}
		out.Position = n.Position

		out.NewTkn = n.NewTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.Class = c.convNode(n.Class)

		if n.Args != nil {
			out.Args = c.convNodeSlice(n.Args)
		} else if hasValue(n.OpenParenthesisTkn) {
			out.Args = []ir.Node{}
		}

		return out

	case *ast.ExprBrackets:
		if n == nil {
			return (*ir.ParenExpr)(nil)
		}
		out := &ir.ParenExpr{}
		out.Position = n.Position
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprPostDec:
		if n == nil {
			return (*ir.PostDecExpr)(nil)
		}
		out := &ir.PostDecExpr{}
		out.Position = n.Position
		out.DecTkn = n.DecTkn
		out.Variable = c.convNode(n.Var)
		return out

	case *ast.ExprPostInc:
		if n == nil {
			return (*ir.PostIncExpr)(nil)
		}
		out := &ir.PostIncExpr{}
		out.Position = n.Position
		out.IncTkn = n.IncTkn
		out.Variable = c.convNode(n.Var)
		return out

	case *ast.ExprPreDec:
		if n == nil {
			return (*ir.PreDecExpr)(nil)
		}
		out := &ir.PreDecExpr{}
		out.Position = n.Position
		out.DecTkn = n.DecTkn
		out.Variable = c.convNode(n.Var)
		return out

	case *ast.ExprPreInc:
		if n == nil {
			return (*ir.PreIncExpr)(nil)
		}
		out := &ir.PreIncExpr{}
		out.Position = n.Position
		out.IncTkn = n.IncTkn
		out.Variable = c.convNode(n.Var)
		return out

	case *ast.ExprPrint:
		if n == nil {
			return (*ir.PrintExpr)(nil)
		}
		out := &ir.PrintExpr{}
		out.Position = n.Position
		out.PrintTkn = n.PrintTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprPropertyFetch:
		if n == nil {
			return (*ir.PropertyFetchExpr)(nil)
		}
		out := &ir.PropertyFetchExpr{}
		out.Position = n.Position
		out.ObjectOperatorTkn = n.ObjectOperatorTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.Variable = c.convNode(n.Var)
		out.Property = c.convNode(n.Prop)
		return out

	case *ast.ExprRequire:
		return c.convImportExpr(n, n.Expr, n.RequireTkn, "require")
	case *ast.ExprRequireOnce:
		return c.convImportExpr(n, n.Expr, n.RequireOnceTkn, "require_once")
	case *ast.ExprInclude:
		return c.convImportExpr(n, n.Expr, n.IncludeTkn, "include")
	case *ast.ExprIncludeOnce:
		return c.convImportExpr(n, n.Expr, n.IncludeOnceTkn, "include_once")

	case *ast.ExprShellExec:
		if n == nil {
			return (*ir.ShellExecExpr)(nil)
		}
		out := &ir.ShellExecExpr{}
		out.Position = n.Position
		out.OpenBacktickTkn = n.OpenBacktickTkn
		out.CloseBacktickTkn = n.CloseBacktickTkn
		out.Parts = c.convNodeSlice(n.Parts)
		return out

	case *ast.ExprStaticCall:
		if n == nil {
			return (*ir.StaticCallExpr)(nil)
		}
		out := &ir.StaticCallExpr{}
		out.Position = n.Position
		out.DoubleColonTkn = n.DoubleColonTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn

		out.Class = c.convNode(n.Class)
		out.Call = c.convNode(n.Call)
		out.Args = c.convNodeSlice(n.Args)
		return out

	case *ast.ExprStaticPropertyFetch:
		if n == nil {
			return (*ir.StaticPropertyFetchExpr)(nil)
		}
		out := &ir.StaticPropertyFetchExpr{}
		out.Position = n.Position
		out.DoubleColonTkn = n.DoubleColonTkn
		out.Class = c.convNode(n.Class)
		out.Property = c.convNode(n.Prop)
		return out

	case *ast.ExprTernary:
		if n == nil {
			return (*ir.TernaryExpr)(nil)
		}
		out := &ir.TernaryExpr{}
		out.Position = n.Position
		out.QuestionTkn = n.QuestionTkn
		out.ColonTkn = n.ColonTkn
		out.Condition = c.convNode(n.Cond)
		out.IfTrue = c.convNode(n.IfTrue)
		out.IfFalse = c.convNode(n.IfFalse)
		return out

	case *ast.ExprUnaryMinus:
		if n == nil {
			return (*ir.UnaryMinusExpr)(nil)
		}
		out := &ir.UnaryMinusExpr{}
		out.Position = n.Position
		out.MinusTkn = n.MinusTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprUnaryPlus:
		if n == nil {
			return (*ir.UnaryPlusExpr)(nil)
		}
		out := &ir.UnaryPlusExpr{}
		out.Position = n.Position
		out.PlusTkn = n.PlusTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.ExprYield:
		if n == nil {
			return (*ir.YieldExpr)(nil)
		}
		out := &ir.YieldExpr{}
		out.Position = n.Position
		out.YieldTkn = n.YieldTkn
		out.DoubleArrowTkn = n.DoubleArrowTkn
		out.Key = c.convNode(n.Key)
		out.Value = c.convNode(n.Val)
		return out

	case *ast.ExprYieldFrom:
		if n == nil {
			return (*ir.YieldFromExpr)(nil)
		}
		out := &ir.YieldFromExpr{}
		out.Position = n.Position
		out.YieldFromTkn = n.YieldFromTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.NameFullyQualified:
		return &ir.Name{
			Position: n.Position,
			Value:    fullyQualifiedToString(n),
		}
	case *ast.Name:
		return &ir.Name{
			Position: n.Position,
			NameTkn:  namePartsToToken(n.Parts),
			Value:    namePartsToString(n.Parts),
		}
	case *ast.NameRelative:
		return c.convRelativeName(n)

	case *ast.Argument:
		if n == nil {
			return (*ir.Argument)(nil)
		}
		out := &ir.Argument{}
		out.Position = n.Position
		out.VariadicTkn = n.VariadicTkn
		out.AmpersandTkn = n.AmpersandTkn
		out.Expr = c.convNode(n.Expr)
		out.Variadic = hasValue(n.VariadicTkn)
		out.IsReference = hasValue(n.AmpersandTkn)
		return out

	case *ast.Identifier:
		if n == nil {
			return (*ir.Identifier)(nil)
		}
		out := &ir.Identifier{}
		out.Position = n.Position
		out.IdentifierTkn = n.IdentifierTkn
		out.Value = string(n.Value)
		return out

	case *ast.Nullable:
		if n == nil {
			return (*ir.Nullable)(nil)
		}
		out := &ir.Nullable{}
		out.Position = n.Position
		out.QuestionTkn = n.QuestionTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.Parameter:
		if n == nil {
			return (*ir.Parameter)(nil)
		}
		out := &ir.Parameter{}
		out.Position = n.Position
		out.AmpersandTkn = n.AmpersandTkn
		out.VariadicTkn = n.VariadicTkn
		out.EqualTkn = n.EqualTkn
		out.VariableType = c.convNode(n.Type)
		out.Variable = c.convNode(n.Var).(*ir.SimpleVar)
		out.DefaultValue = c.convNode(n.DefaultValue)

		out.ByRef = hasValue(n.AmpersandTkn)
		out.Variadic = hasValue(n.VariadicTkn)
		return out

	case *ast.Root:
		if n == nil {
			return (*ir.Root)(nil)
		}
		out := &ir.Root{}
		out.Position = n.Position
		out.EndTkn = n.EndTkn
		{
			slice := make([]ir.Node, len(n.Stmts))
			for i := range n.Stmts {
				slice[i] = c.convNode(n.Stmts[i])
			}
			out.Stmts = slice
		}
		return out

	case *ast.ExprVariable:
		if n == nil {
			return (*ir.SimpleVar)(nil)
		}

		nameNode, ok := n.Name.(*ast.Identifier)
		if !ok {
			return &ir.Var{
				Position:             n.Position,
				DollarTkn:            n.DollarTkn,
				OpenCurlyBracketTkn:  n.OpenCurlyBracketTkn,
				Expr:                 c.convNode(n.Name),
				CloseCurlyBracketTkn: n.CloseCurlyBracketTkn,
			}
		}

		nameNodeIr := c.convNode(nameNode).(*ir.Identifier)

		out := &ir.SimpleVar{}
		out.Position = n.Position
		out.DollarTkn = n.DollarTkn
		out.Name = string(bytes.TrimPrefix(nameNode.Value, []byte("$")))
		out.IdentifierTkn = nameNodeIr.IdentifierTkn

		return out

	case *ast.ScalarDnumber:
		if n == nil {
			return (*ir.Dnumber)(nil)
		}
		out := &ir.Dnumber{}
		out.Position = n.Position
		out.NumberTkn = n.NumberTkn
		out.Value = string(n.Value)
		return out

	case *ast.ScalarEncapsed:
		if n == nil {
			return (*ir.Encapsed)(nil)
		}
		out := &ir.Encapsed{}
		out.Position = n.Position
		out.OpenQuoteTkn = n.OpenQuoteTkn
		out.CloseQuoteTkn = n.CloseQuoteTkn
		out.Parts = c.convNodeSlice(n.Parts)
		return out

	case *ast.ScalarEncapsedStringBrackets:
		if n == nil {
			return nil
		}

		return c.convNode(n.Var)

	case *ast.ScalarEncapsedStringVar:
		if n == nil {
			return nil
		}

		if n.Dim != nil {
			return &ir.ArrayDimFetchExpr{
				Position:        n.Position,
				Variable:        c.convNode(n.Name),
				OpenBracketTkn:  n.OpenSquareBracketTkn,
				Dim:             c.convNode(n.Dim),
				CloseBracketTkn: n.CloseSquareBracketTkn,
			}
		}

		nameNode := c.ConvertNode(n.Name).(*ir.Identifier)

		return &ir.SimpleVar{
			Position:      n.Position,
			IdentifierTkn: nameNode.IdentifierTkn,
			Name:          nameNode.Value,
		}

	case *ast.ScalarEncapsedStringPart:
		if n == nil {
			return (*ir.EncapsedStringPart)(nil)
		}
		out := &ir.EncapsedStringPart{}
		out.Position = n.Position
		out.EncapsedStrTkn = n.EncapsedStrTkn
		out.Value = string(n.Value)
		return out

	case *ast.ScalarHeredoc:
		if n == nil {
			return (*ir.Heredoc)(nil)
		}
		out := &ir.Heredoc{}
		out.Position = n.Position
		out.OpenHeredocTkn = n.OpenHeredocTkn
		out.CloseHeredocTkn = n.CloseHeredocTkn
		out.Label = string(n.OpenHeredocTkn.Value)
		out.Parts = c.convNodeSlice(n.Parts)
		return out

	case *ast.ScalarLnumber:
		if n == nil {
			return (*ir.Lnumber)(nil)
		}
		out := &ir.Lnumber{}
		out.Position = n.Position
		out.NumberTkn = n.NumberTkn
		out.Value = string(n.Value)
		return out

	case *ast.ScalarMagicConstant:
		if n == nil {
			return (*ir.MagicConstant)(nil)
		}
		out := &ir.MagicConstant{}
		out.Position = n.Position
		out.MagicConstTkn = n.MagicConstTkn
		out.Value = string(n.Value)
		return out

	case *ast.ScalarString:
		return convString(n)

	case *ast.StmtBreak:
		if n == nil {
			return (*ir.BreakStmt)(nil)
		}
		out := &ir.BreakStmt{}
		out.Position = n.Position
		out.BreakTkn = n.BreakTkn
		out.SemiColonTkn = n.SemiColonTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.StmtCase:
		if n == nil {
			return (*ir.CaseStmt)(nil)
		}
		out := &ir.CaseStmt{}
		out.Position = n.Position
		out.CaseTkn = n.CaseTkn
		out.CaseSeparatorTkn = n.CaseSeparatorTkn
		out.Cond = c.convNode(n.Cond)
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.StmtCatch:
		if n == nil {
			return (*ir.CatchStmt)(nil)
		}
		out := &ir.CatchStmt{}
		out.Position = n.Position

		out.CatchTkn = n.CatchTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn

		out.Types = c.convNodeSlice(n.Types)
		out.Variable = c.convNode(n.Var).(*ir.SimpleVar)
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.StmtClass:
		if n == nil {
			return (*ir.ClassStmt)(nil)
		}
		return c.convClass(n)

	case *ast.StmtClassConstList:
		if n == nil {
			return (*ir.ClassConstListStmt)(nil)
		}
		out := &ir.ClassConstListStmt{}
		out.Position = n.Position
		out.ConstTkn = n.ConstTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.SemiColonTkn = n.SemiColonTkn
		{
			slice := make([]*ir.Identifier, len(n.Modifiers))
			for i := range n.Modifiers {
				slice[i] = c.convNode(n.Modifiers[i]).(*ir.Identifier)
			}
			out.Modifiers = slice
		}

		out.PhpDocComment, out.PhpDoc = c.getPhpDocWithParse(n.ConstTkn)

		out.Consts = c.convNodeSlice(n.Consts)
		return out

	case *ast.StmtClassMethod:
		if n == nil {
			return (*ir.ClassMethodStmt)(nil)
		}
		out := &ir.ClassMethodStmt{}
		out.Position = n.Position

		out.FunctionTkn = n.FunctionTkn
		out.AmpersandTkn = n.AmpersandTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn

		var tokenWithDoc *token.Token
		if len(n.Modifiers) != 0 {
			tokenWithDoc = n.Modifiers[0].(*ast.Identifier).IdentifierTkn
		} else {
			tokenWithDoc = n.FunctionTkn
		}

		out.PhpDocComment, out.PhpDoc = c.getPhpDocWithParse(tokenWithDoc)

		out.MethodName = c.convNode(n.Name).(*ir.Identifier)
		{
			slice := make([]*ir.Identifier, len(n.Modifiers))
			for i := range n.Modifiers {
				slice[i] = c.convNode(n.Modifiers[i]).(*ir.Identifier)
			}
			out.Modifiers = slice
		}
		out.Params = c.convNodeSlice(n.Params)
		out.ReturnType = c.convNode(n.ReturnType)
		out.Stmt = c.convNode(n.Stmt)
		out.ReturnsRef = hasValue(n.AmpersandTkn)
		return out

	case *ast.StmtConstList:
		if n == nil {
			return (*ir.ConstListStmt)(nil)
		}
		out := &ir.ConstListStmt{}
		out.Position = n.Position
		out.ConstTkn = n.ConstTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.SemiColonTkn = n.SemiColonTkn
		out.Consts = c.convNodeSlice(n.Consts)
		return out

	case *ast.StmtConstant:
		if n == nil {
			return (*ir.ConstantStmt)(nil)
		}
		out := &ir.ConstantStmt{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.ConstantName = c.convNode(n.Name).(*ir.Identifier)
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.StmtContinue:
		if n == nil {
			return (*ir.ContinueStmt)(nil)
		}
		out := &ir.ContinueStmt{}
		out.Position = n.Position
		out.ContinueTkn = n.ContinueTkn
		out.SemiColonTkn = n.SemiColonTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.StmtDeclare:
		if n == nil {
			return (*ir.DeclareStmt)(nil)
		}
		out := &ir.DeclareStmt{}
		out.Position = n.Position
		out.DeclareTkn = n.DeclareTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn
		out.EndDeclareTkn = n.EndDeclareTkn

		out.Consts = c.convNodeSlice(n.Consts)
		out.Stmt = c.convNode(n.Stmt)
		out.Alt = hasValue(n.EndDeclareTkn)
		return out

	case *ast.StmtDefault:
		if n == nil {
			return (*ir.DefaultStmt)(nil)
		}
		out := &ir.DefaultStmt{}
		out.Position = n.Position
		out.DefaultTkn = n.DefaultTkn
		out.CaseSeparatorTkn = n.CaseSeparatorTkn
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.StmtDo:
		if n == nil {
			return (*ir.DoStmt)(nil)
		}
		out := &ir.DoStmt{}
		out.Position = n.Position

		out.DoTkn = n.DoTkn
		out.WhileTkn = n.WhileTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.SemiColonTkn = n.SemiColonTkn

		out.Stmt = c.convNode(n.Stmt)
		out.Cond = c.convNode(n.Cond)
		return out

	case *ast.StmtEcho:
		if n == nil {
			return (*ir.EchoStmt)(nil)
		}
		out := &ir.EchoStmt{}
		out.Position = n.Position
		out.EchoTkn = n.EchoTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.SemiColonTkn = n.SemiColonTkn
		out.Exprs = c.convNodeSlice(n.Exprs)
		return out

	case *ast.StmtElse:
		if n == nil {
			return (*ir.ElseStmt)(nil)
		}
		out := &ir.ElseStmt{}
		out.Position = n.Position
		out.ElseTkn = n.ElseTkn
		out.ColonTkn = n.ColonTkn
		out.AltSyntax = hasValue(n.ColonTkn)

		out.Stmt = c.convNode(n.Stmt)

		// Since the parser turns the else if statement into an ir.ElseStmt
		// node with the Stmt field equal to the ir.IfStmt node, we need
		// to convert this to an ir.ElseIfStmt node.
		// For this, if ir.ElseStmt contains ir.IfStmt, then it is necessary to
		// return the ir.IfStmt node, which contains the necessary fields,
		// to create the ir.ElseIfStmt node in the future.
		if ifStmt, ok := out.Stmt.(*ir.IfStmt); ok {
			ifStmt.ElseTkn = n.ElseTkn
			return ifStmt
		}

		return out

	case *ast.StmtElseIf:
		if n == nil {
			return (*ir.ElseIfStmt)(nil)
		}
		out := &ir.ElseIfStmt{}
		out.ElseIfTkn = n.ElseIfTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn

		out.Position = n.Position
		out.Cond = c.convNode(n.Cond)
		out.Stmt = c.convNode(n.Stmt)

		out.AltSyntax = hasValue(n.ColonTkn)

		// directly converting ElseIf always means they are merged
		out.Merged = true
		return out

	case *ast.StmtExpression:
		if n == nil {
			return (*ir.ExpressionStmt)(nil)
		}
		out := &ir.ExpressionStmt{}
		out.Position = n.Position
		out.SemiColonTkn = n.SemiColonTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.StmtFinally:
		if n == nil {
			return (*ir.FinallyStmt)(nil)
		}
		out := &ir.FinallyStmt{}
		out.Position = n.Position
		out.FinallyTkn = n.FinallyTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.StmtFor:
		if n == nil {
			return (*ir.ForStmt)(nil)
		}
		out := &ir.ForStmt{}
		out.Position = n.Position

		out.ForTkn = n.ForTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.InitSeparatorTkns = n.InitSeparatorTkns
		out.InitSemiColonTkn = n.InitSemiColonTkn
		out.CondSeparatorTkns = n.CondSeparatorTkns
		out.CondSemiColonTkn = n.CondSemiColonTkn
		out.LoopSeparatorTkns = n.LoopSeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn
		out.EndForTkn = n.EndForTkn
		out.SemiColonTkn = n.SemiColonTkn

		out.Init = c.convNodeSlice(n.Init)
		out.Cond = c.convNodeSlice(n.Cond)
		out.Loop = c.convNodeSlice(n.Loop)
		out.Stmt = c.convNode(n.Stmt)

		out.AltSyntax = hasValue(n.EndForTkn)
		return out

	case *ast.StmtForeach:
		if n == nil {
			return (*ir.ForeachStmt)(nil)
		}
		out := &ir.ForeachStmt{}
		out.Position = n.Position

		out.ForeachTkn = n.ForeachTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.AsTkn = n.AsTkn
		out.DoubleArrowTkn = n.DoubleArrowTkn
		out.AmpersandTkn = n.AmpersandTkn

		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn
		out.EndForeachTkn = n.EndForeachTkn
		out.SemiColonTkn = n.SemiColonTkn

		out.Expr = c.convNode(n.Expr)
		out.Key = c.convNode(n.Key)

		if hasValue(n.AmpersandTkn) {
			out.Variable = &ir.ReferenceExpr{
				FreeFloating: nil,
				AmpersandTkn: n.AmpersandTkn,
				Position:     n.Position,
				Variable:     c.convNode(n.Var),
			}
		} else {
			out.Variable = c.convNode(n.Var)
		}

		out.Stmt = c.convNode(n.Stmt)

		out.AltSyntax = hasValue(n.EndForeachTkn)
		return out

	case *ast.StmtFunction:
		if n == nil {
			return (*ir.FunctionStmt)(nil)
		}
		out := &ir.FunctionStmt{}
		out.Position = n.Position

		out.FunctionTkn = n.FunctionTkn
		out.AmpersandTkn = n.AmpersandTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn

		out.PhpDocComment, out.PhpDoc = c.getPhpDocWithParse(n.FunctionTkn)

		out.FunctionName = c.convNode(n.Name).(*ir.Identifier)
		out.Params = c.convNodeSlice(n.Params)
		out.ReturnType = c.convNode(n.ReturnType)
		out.Stmts = c.convNodeSlice(n.Stmts)

		out.ReturnsRef = hasValue(n.AmpersandTkn)
		return out

	case *ast.StmtGlobal:
		if n == nil {
			return (*ir.GlobalStmt)(nil)
		}
		out := &ir.GlobalStmt{}
		out.Position = n.Position
		out.GlobalTkn = n.GlobalTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.SemiColonTkn = n.SemiColonTkn
		out.Vars = c.convNodeSlice(n.Vars)
		return out

	case *ast.StmtGoto:
		if n == nil {
			return (*ir.GotoStmt)(nil)
		}
		out := &ir.GotoStmt{}
		out.Position = n.Position
		out.GotoTkn = n.GotoTkn
		out.SemiColonTkn = n.SemiColonTkn
		out.Label = c.convNode(n.Label).(*ir.Identifier)
		return out

	case *ast.StmtGroupUseList:
		if n == nil {
			return (*ir.GroupUseStmt)(nil)
		}
		out := &ir.GroupUseStmt{}
		out.Position = n.Position

		out.UseTkn = n.UseTkn
		out.LeadingNsSeparatorTkn = n.LeadingNsSeparatorTkn
		out.NsSeparatorTkn = n.NsSeparatorTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.SemiColonTkn = n.SemiColonTkn

		useType := c.convNode(n.Type)
		if useType != nil {
			out.UseType = useType.(*ir.Identifier)
		}
		out.Prefix = c.convNode(n.Prefix).(*ir.Name)
		out.UseList = c.convNodeSlice(n.Uses)
		return out

	case *ast.StmtHaltCompiler:
		if n == nil {
			return (*ir.HaltCompilerStmt)(nil)
		}
		out := &ir.HaltCompilerStmt{}
		out.Position = n.Position
		out.HaltCompilerTkn = n.HaltCompilerTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.SemiColonTkn = n.SemiColonTkn
		return out

	case *ast.StmtIf:
		if n == nil {
			return (*ir.IfStmt)(nil)
		}
		out := &ir.IfStmt{}
		out.Position = n.Position

		out.IfTkn = n.IfTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn
		out.EndIfTkn = n.EndIfTkn
		out.SemiColonTkn = n.SemiColonTkn

		out.Cond = c.convNode(n.Cond)
		out.Stmt = c.convNode(n.Stmt)
		out.ElseIf = c.convNodeSlice(n.ElseIf)
		out.Else = c.convNode(n.Else)

		// Since the parser convert the else if statement into
		// an ir.ElseStmt node with the Stmt field equal to the
		// ir.IfStmt node, we need to convert this to an ir.ElseIfStmt node.
		//
		// For this, if ir.ElseStmt contains ir.IfStmt, then the converter returns
		// the ir.IfStmt node, which contains the necessary fields to create
		// the ir.ElseIfStmt node.
		if ifStmt, ok := out.Else.(*ir.IfStmt); ok {
			ifStmt.Position.StartPos = n.Position.StartPos
			ifStmt.Position.StartLine = n.Position.StartLine

			out.ElseIf = append(out.ElseIf, &ir.ElseIfStmt{
				Position:            ifStmt.Position,
				IfTkn:               ifStmt.IfTkn,
				ElseTkn:             ifStmt.ElseTkn,
				OpenParenthesisTkn:  ifStmt.OpenParenthesisTkn,
				Cond:                ifStmt.Cond,
				CloseParenthesisTkn: ifStmt.CloseParenthesisTkn,
				ColonTkn:            ifStmt.ColonTkn,
				Stmt:                ifStmt.Stmt,
				AltSyntax:           ifStmt.AltSyntax,
				Merged:              false,
			})

			out.ElseIf = append(out.ElseIf, ifStmt.ElseIf...)

			out.Else = ifStmt.Else
		}

		out.AltSyntax = hasValue(n.ColonTkn)
		return out

	case *ast.StmtInlineHtml:
		if n == nil {
			return (*ir.InlineHTMLStmt)(nil)
		}
		out := &ir.InlineHTMLStmt{}
		out.Position = n.Position
		out.InlineHTMLTkn = n.InlineHtmlTkn
		out.Value = string(n.Value)
		return out

	case *ast.StmtInterface:
		if n == nil {
			return (*ir.InterfaceStmt)(nil)
		}
		out := &ir.InterfaceStmt{}
		out.Position = n.Position

		out.InterfaceTkn = n.InterfaceTkn
		out.ExtendsTkn = n.ExtendsTkn
		out.ExtendsSeparatorTkns = n.ExtendsSeparatorTkns
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn

		out.PhpDocComment, out.PhpDoc = c.getPhpDocWithParse(n.InterfaceTkn)

		out.InterfaceName = c.convNode(n.Name).(*ir.Identifier)
		out.Extends = &ir.InterfaceExtendsStmt{
			InterfaceNames: c.convNodeSlice(n.Extends),
		}
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.StmtLabel:
		if n == nil {
			return (*ir.LabelStmt)(nil)
		}
		out := &ir.LabelStmt{}
		out.Position = n.Position
		out.ColonTkn = n.ColonTkn
		out.LabelName = c.convNode(n.Name).(*ir.Identifier)
		return out

	case *ast.StmtNamespace:
		if n == nil {
			return (*ir.NamespaceStmt)(nil)
		}
		out := &ir.NamespaceStmt{}
		out.Position = n.Position

		out.NsTkn = n.NsTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.SemiColonTkn = n.SemiColonTkn
		if n.Name != nil {
			out.NamespaceName = c.convNode(n.Name).(*ir.Name)
			c.namespace = out.NamespaceName.Value
		}
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.StmtNop:
		if n == nil {
			return (*ir.NopStmt)(nil)
		}
		out := &ir.NopStmt{}
		out.Position = n.Position
		out.SemiColonTkn = n.SemiColonTkn
		return out

	case *ast.StmtProperty:
		if n == nil {
			return (*ir.PropertyStmt)(nil)
		}
		out := &ir.PropertyStmt{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var).(*ir.SimpleVar)
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.StmtPropertyList:
		if n == nil {
			return (*ir.PropertyListStmt)(nil)
		}
		out := &ir.PropertyListStmt{}
		out.Position = n.Position
		out.SeparatorTkns = n.SeparatorTkns
		out.SemiColonTkn = n.SemiColonTkn
		{
			slice := make([]*ir.Identifier, len(n.Modifiers))
			for i := range n.Modifiers {
				slice[i] = c.convNode(n.Modifiers[i]).(*ir.Identifier)
			}
			out.Modifiers = slice
		}

		var tokenWithDoc *token.Token
		if len(n.Modifiers) != 0 {
			tokenWithDoc = n.Modifiers[0].(*ast.Identifier).IdentifierTkn
		}
		out.PhpDocComment, out.PhpDoc = c.getPhpDocWithParse(tokenWithDoc)

		out.Type = c.convNode(n.Type)
		out.Properties = c.convNodeSlice(n.Props)
		return out

	case *ast.StmtReturn:
		if n == nil {
			return (*ir.ReturnStmt)(nil)
		}
		out := &ir.ReturnStmt{}
		out.Position = n.Position
		out.ReturnTkn = n.ReturnTkn
		out.SemiColonTkn = n.SemiColonTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.StmtStatic:
		if n == nil {
			return (*ir.StaticStmt)(nil)
		}
		out := &ir.StaticStmt{}
		out.Position = n.Position
		out.StaticTkn = n.StaticTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.Vars = c.convNodeSlice(n.Vars)
		return out

	case *ast.StmtStaticVar:
		if n == nil {
			return (*ir.StaticVarStmt)(nil)
		}
		out := &ir.StaticVarStmt{}
		out.Position = n.Position
		out.EqualTkn = n.EqualTkn
		out.Variable = c.convNode(n.Var).(*ir.SimpleVar)
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.StmtStmtList:
		if n == nil {
			return (*ir.StmtList)(nil)
		}
		out := &ir.StmtList{}
		out.Position = n.Position
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.StmtSwitch:
		if n == nil {
			return (*ir.SwitchStmt)(nil)
		}
		out := &ir.SwitchStmt{}
		out.Position = n.Position

		out.SwitchTkn = n.SwitchTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CaseSeparatorTkn = n.CaseSeparatorTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.EndSwitchTkn = n.EndSwitchTkn
		out.SemiColonTkn = n.SemiColonTkn

		out.Cond = c.convNode(n.Cond)
		out.Cases = c.convNodeSlice(n.Cases)
		out.AltSyntax = hasValue(n.ColonTkn)
		return out

	case *ast.StmtThrow:
		if n == nil {
			return (*ir.ThrowStmt)(nil)
		}
		out := &ir.ThrowStmt{}
		out.Position = n.Position
		out.ThrowTkn = n.ThrowTkn
		out.SemiColonTkn = n.SemiColonTkn
		out.Expr = c.convNode(n.Expr)
		return out

	case *ast.StmtTrait:
		if n == nil {
			return (*ir.TraitStmt)(nil)
		}
		out := &ir.TraitStmt{}
		out.Position = n.Position

		out.TraitTkn = n.TraitTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn

		out.PhpDocComment, out.PhpDoc = c.getPhpDocWithParse(out.TraitTkn)

		out.TraitName = c.convNode(n.Name).(*ir.Identifier)
		out.Stmts = c.convNodeSlice(n.Stmts)
		return out

	case *ast.StmtTraitUse:
		if n == nil {
			return (*ir.TraitUseStmt)(nil)
		}
		out := &ir.TraitUseStmt{}
		out.Position = n.Position

		out.UseTkn = n.UseTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.SemiColonTkn = n.SemiColonTkn

		out.Traits = c.convNodeSlice(n.Traits)
		// TODO:
		out.TraitAdaptationList = &ir.TraitAdaptationListStmt{
			Adaptations: c.convNodeSlice(n.Adaptations),
		}
		return out

	case *ast.StmtTraitUseAlias:
		if n == nil {
			return (*ir.TraitUseAliasStmt)(nil)
		}
		out := &ir.TraitUseAliasStmt{}
		out.Position = n.Position
		out.DoubleColonTkn = n.DoubleColonTkn
		out.AsTkn = n.AsTkn
		out.SemiColonTkn = n.SemiColonTkn
		out.Ref = c.convNode(n.Method)
		out.Modifier = c.convNode(n.Modifier)
		out.Alias = c.convNode(n.Alias).(*ir.Identifier)
		return out

	case *ast.StmtTraitUsePrecedence:
		if n == nil {
			return (*ir.TraitUsePrecedenceStmt)(nil)
		}
		out := &ir.TraitUsePrecedenceStmt{}
		out.Position = n.Position
		out.DoubleColonTkn = n.DoubleColonTkn
		out.InsteadofTkn = n.InsteadofTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.SemiColonTkn = n.SemiColonTkn
		out.Ref = c.convNode(n.Method)
		out.Insteadof = c.convNodeSlice(n.Insteadof)
		return out

	case *ast.StmtTry:
		if n == nil {
			return (*ir.TryStmt)(nil)
		}
		out := &ir.TryStmt{}
		out.Position = n.Position
		out.TryTkn = n.TryTkn
		out.OpenCurlyBracketTkn = n.OpenCurlyBracketTkn
		out.CloseCurlyBracketTkn = n.CloseCurlyBracketTkn
		out.Stmts = c.convNodeSlice(n.Stmts)
		out.Catches = c.convNodeSlice(n.Catches)
		out.Finally = c.convNode(n.Finally)
		return out

	case *ast.StmtUnset:
		if n == nil {
			return (*ir.UnsetStmt)(nil)
		}
		out := &ir.UnsetStmt{}
		out.Position = n.Position
		out.UnsetTkn = n.UnsetTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.SemiColonTkn = n.SemiColonTkn
		out.Vars = c.convNodeSlice(n.Vars)
		return out

	case *ast.StmtUse:
		if n == nil {
			return (*ir.UseStmt)(nil)
		}
		out := &ir.UseStmt{}
		out.Position = n.Position
		out.NsSeparatorTkn = n.NsSeparatorTkn
		out.AsTkn = n.AsTkn
		if n.Type != nil {
			out.UseType = c.convNode(n.Type).(*ir.Identifier)
		}
		out.Use = c.convNode(n.Use).(*ir.Name)
		if n.Alias != nil {
			out.Alias = c.convNode(n.Alias).(*ir.Identifier)
		}
		return out

	case *ast.StmtUseList:
		if n == nil {
			return (*ir.UseListStmt)(nil)
		}
		out := &ir.UseListStmt{}
		out.Position = n.Position
		out.UseTkn = n.UseTkn
		out.SeparatorTkns = n.SeparatorTkns
		out.SemiColonTkn = n.SemiColonTkn
		useType := c.convNode(n.Type)
		if useType != nil {
			out.UseType = useType.(*ir.Identifier)
		}
		out.Uses = c.convNodeSlice(n.Uses)
		return out

	case *ast.StmtWhile:
		if n == nil {
			return (*ir.WhileStmt)(nil)
		}
		out := &ir.WhileStmt{}
		out.Position = n.Position

		out.WhileTkn = n.WhileTkn
		out.OpenParenthesisTkn = n.OpenParenthesisTkn
		out.CloseParenthesisTkn = n.CloseParenthesisTkn
		out.ColonTkn = n.ColonTkn
		out.EndWhileTkn = n.EndWhileTkn
		out.SemiColonTkn = n.SemiColonTkn

		out.Cond = c.convNode(n.Cond)
		out.Stmt = c.convNode(n.Stmt)
		out.AltSyntax = hasValue(n.EndWhileTkn)
		return out
	}

	panic(fmt.Sprintf("unhandled type %T", n))
}

func hasValue(tok *token.Token) bool {
	return tok != nil
}

func (c *Converter) getPhpDocWithParse(tok *token.Token) (doc string, parsed []phpdoc.CommentPart) {
	if tok == nil {
		return doc, parsed
	}

	for _, ff := range tok.FreeFloating {
		if ff.ID == token.T_DOC_COMMENT {
			doc = string(ff.Value)
			parsed = c.parsePHPDoc(doc)
		}
	}

	return doc, parsed
}

func (c *Converter) convRelativeName(n *ast.NameRelative) *ir.Name {
	value := namePartsToString(n.Parts)
	if c.namespace != "" {
		value = `\` + c.namespace + `\` + value
	}
	return &ir.Name{
		Position: n.Position,
		Value:    value,
	}
}

func (c *Converter) convImportExpr(n, e ast.Vertex, importTkn *token.Token, fn string) *ir.ImportExpr {
	return &ir.ImportExpr{
		ImportTkn: importTkn,
		Position:  n.GetPosition(),
		Func:      fn,
		Expr:      c.convNode(e),
	}
}

func (c *Converter) convCastExpr(n, e ast.Vertex, castTkn *token.Token, typ string) *ir.TypeCastExpr {
	return &ir.TypeCastExpr{
		CastTkn:  castTkn,
		Position: n.GetPosition(),
		Type:     typ,
		Expr:     c.convNode(e),
	}
}

func (c *Converter) convClass(n *ast.StmtClass) ir.Node {
	var extends *ir.ClassExtendsStmt
	extendsNode := c.convNode(n.Extends)
	if extendsNode != nil {
		extends = &ir.ClassExtendsStmt{
			Position:   n.ExtendsTkn.Position,
			ExtendsTkn: n.ExtendsTkn,
			ClassName:  extendsNode.(*ir.Name),
		}
	}

	class := ir.Class{
		Extends: extends,
		Stmts:   c.convNodeSlice(n.Stmts),
	}

	implements := c.convNodeSlice(n.Implements)

	if len(implements) != 0 {
		class.Implements = &ir.ClassImplementsStmt{
			Position:                n.ImplementsTkn.Position,
			ImplementsTkn:           n.ImplementsTkn,
			ImplementsSeparatorTkns: n.ImplementsSeparatorTkns,
			InterfaceNames:          implements,
		}
	}

	class.PhpDocComment, class.PhpDoc = c.getPhpDocWithParse(n.ClassTkn)

	if n.Name == nil {
		// Anonymous class expression.
		out := &ir.AnonClassExpr{
			ClassTkn:             n.ClassTkn,
			OpenParenthesisTkn:   n.OpenParenthesisTkn,
			SeparatorTkns:        n.SeparatorTkns,
			CloseParenthesisTkn:  n.CloseParenthesisTkn,
			OpenCurlyBracketTkn:  n.OpenCurlyBracketTkn,
			CloseCurlyBracketTkn: n.CloseCurlyBracketTkn,
			Position:             n.Position,
			Class:                class,
		}
		if n.Args != nil {
			out.Args = c.convNodeSlice(n.Args)
		}
		return out
	}

	out := &ir.ClassStmt{
		ClassTkn:             n.ClassTkn,
		OpenCurlyBracketTkn:  n.OpenCurlyBracketTkn,
		CloseCurlyBracketTkn: n.CloseCurlyBracketTkn,
		Position:             n.Position,
		Class:                class,
		ClassName:            c.convNode(n.Name).(*ir.Identifier),
	}
	if n.Modifiers != nil {
		slice := make([]*ir.Identifier, len(n.Modifiers))
		for i := range n.Modifiers {
			slice[i] = c.convNode(n.Modifiers[i]).(*ir.Identifier)
		}
		out.Modifiers = slice
	}
	return out
}

func (c *Converter) parsePHPDoc(doc string) []phpdoc.CommentPart {
	if c.phpdocTypeParser != nil {
		return phpdoc.Parse(c.phpdocTypeParser, doc)
	}
	return nil
}

func convString(n *ast.ScalarString) ir.Node {
	out := &ir.String{
		MinusTkn:  n.MinusTkn,
		StringTkn: n.StringTkn,
		Position:  n.Position,
	}

	// We can't use n.Value[0] as quote char directly as when
	// we parse string parts like $_SERVER[HTTP_HOST] we get
	// HTTP_HOST as a value with no quotes.
	var quote byte
	if n.Value[0] == '"' {
		quote = '"'
	} else {
		quote = '\''
	}

	out.DoubleQuotes = n.Value[0] == '"'
	unquoted := irutil.Unquote(string(n.Value))
	s, err := interpretString(unquoted, quote)
	if err != nil {
		return &ir.BadString{
			MinusTkn:     n.MinusTkn,
			StringTkn:    n.StringTkn,
			Position:     n.Position,
			Value:        unquoted,
			Error:        err.Error(),
			DoubleQuotes: out.DoubleQuotes,
		}
	}
	out.Value = s

	return out
}
