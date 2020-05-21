package rules

import (
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/binary"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/cast"
	"github.com/VKCOM/noverify/src/php/parser/node/scalar"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
)

type RuleKind int

// All kinds of node categories.
//
// See CategorizeNode to see how they are connected
// with particular node types.
const (
	KindNone RuleKind = iota

	KindMethodCall
	KindFunctionCall
	KindStaticCall
	KindArray
	KindTernary

	KindValueFetch
	KindCmp     // All binary comparison ops
	KindBinOp   // All binary ops except comparison ops
	KindUnaryOp // All unary-ish ops
	KindAssign  // All assignments
	KindCondOp  // All conditional expressions
	KindBranch  // All conditional statements
	KindScalar  // All scalar-like nodes
	KindConst   // All const fetching ops
	KindRequire
	KindLoop
	KindCast
	KindOther         // All remaining kinds that are not None
	KindOtherUnlikely // Second Other category, even less priority

	_KindCount // Should be always the last one
)

// CategorizeNode tries to associate a node with one of the known categories.
// If node can't be (yet?) categorized, returns KindNone.
//
// Note that for some nodes we *want* to return KindNone, since matching
// engine can have a hard time trying to match them.
//
// We try to categorize nodes in a way that predict which rules have
// higher chances to be used by the users. For examples, it's likely
// that function calls could be the most important pattern out there.
// With some other groups it's more or less shot in the dark.
// A better algorithm can be found later. Maybe we'll assign categories
// dynamically or map them 1-to-1 with nodes kind in future.
func CategorizeNode(n node.Node) RuleKind {
	switch n.(type) {
	case *expr.FunctionCall:
		return KindFunctionCall

	case *expr.MethodCall:
		return KindMethodCall

	case *expr.StaticCall:
		return KindStaticCall

	case *expr.Array:
		return KindArray

	case *expr.PropertyFetch,
		*expr.StaticPropertyFetch,
		*expr.ArrayDimFetch:
		return KindValueFetch

	case *expr.Isset,
		*expr.Empty,
		*stmt.Return,
		*expr.New:
		return KindOther

	case *expr.Exit,
		*expr.Yield,
		*expr.Eval,
		*expr.Print,
		*expr.Clone:
		return KindOtherUnlikely

	case *assign.Assign,
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
		*assign.ShiftRight:
		return KindAssign

	case *binary.Equal,
		*binary.NotEqual,
		*binary.Identical,
		*binary.NotIdentical,
		*binary.Smaller,
		*binary.Greater,
		*binary.SmallerOrEqual,
		*binary.GreaterOrEqual,
		*binary.Spaceship:
		return KindCmp

	case *binary.BitwiseAnd,
		*binary.BitwiseOr,
		*binary.BitwiseXor,
		*binary.Coalesce,
		*binary.Concat,
		*binary.Div,
		*binary.Minus,
		*binary.Mod,
		*binary.Mul,
		*binary.Plus,
		*binary.Pow,
		*binary.ShiftLeft,
		*binary.ShiftRight:
		return KindBinOp

	case *expr.BitwiseNot,
		*expr.BooleanNot,
		*expr.ErrorSuppress,
		*expr.PostDec,
		*expr.PostInc,
		*expr.PreDec,
		*expr.PreInc,
		*expr.Reference,
		*expr.UnaryMinus:
		return KindUnaryOp

	case *binary.LogicalOr,
		*binary.LogicalAnd,
		*binary.BooleanOr,
		*binary.BooleanAnd:
		return KindCondOp

	case *expr.Ternary:
		return KindTernary

	case *stmt.Do,
		*stmt.For,
		*stmt.Foreach,
		*stmt.While:
		return KindLoop

	case *cast.Array,
		*cast.Bool,
		*cast.Double,
		*cast.Int,
		*cast.Object,
		*cast.String,
		*cast.Unset:
		return KindCast

	case *expr.Require,
		*expr.RequireOnce,
		*expr.Include,
		*expr.IncludeOnce:
		return KindRequire

	case *stmt.If,
		*stmt.Throw:
		return KindBranch

	case *expr.ShellExec,
		*scalar.String,
		*scalar.Lnumber,
		*scalar.Dnumber,
		*scalar.Heredoc,
		*scalar.Encapsed:
		return KindScalar

	case *expr.ClassConstFetch,
		*expr.ConstFetch:
		return KindConst

	default:
		return KindNone
	}
}
