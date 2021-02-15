package irutil

import (
	"github.com/z7zmey/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irfmt"
)

// Unquote returns unquoted version of s, if there are any quotes.
func Unquote(s string) string {
	if len(s) >= 2 && s[0] == '\'' || s[0] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func NodeSliceClone(xs []ir.Node) []ir.Node {
	cloned := make([]ir.Node, len(xs))
	for i, x := range xs {
		cloned[i] = NodeClone(x)
	}
	return cloned
}

// Unparen returns n with all parenthesis removed.
func Unparen(e ir.Node) ir.Node {
	for {
		p, ok := e.(*ir.ParenExpr)
		if !ok {
			return e
		}
		e = p.Expr
	}
}

func NodeSliceEqual(xs, ys []ir.Node) bool {
	if len(xs) != len(ys) {
		return false
	}
	for i, x := range xs {
		if !NodeEqual(x, ys[i]) {
			return false
		}
	}
	return true
}

func IsAssign(n ir.Node) bool {
	switch n.(type) {
	case *ir.Assign,
		*ir.AssignConcat,
		*ir.AssignPlus,
		*ir.AssignReference,
		*ir.AssignDiv,
		*ir.AssignPow,
		*ir.AssignBitwiseAnd,
		*ir.AssignBitwiseOr,
		*ir.AssignBitwiseXor,
		*ir.AssignShiftLeft,
		*ir.AssignShiftRight,
		*ir.AssignMinus,
		*ir.AssignMod,
		*ir.AssignMul,
		*ir.AssignCoalesce:
		return true
	default:
		return false
	}
}

// FmtNode returns string representation of n.
func FmtNode(n ir.Node) string {
	return irfmt.Node(n)
}

// FindPhpDoc searches for phpdoc by traversing all subtree and all tokens.
func FindPhpDoc(n ir.Node) (string, bool) {
	var doc string

	Inspect(n, func(n ir.Node) bool {
		n.IterateTokens(func(t *token.Token) bool {
			if t.ID == token.T_DOC_COMMENT {
				doc = string(t.Value)
				return false
			}

			return doc == ""
		})

		return doc == ""
	})

	return doc, doc != ""
}

func classEqual(x, y ir.Class) bool {
	return x.PhpDocComment == y.PhpDocComment &&
		NodeEqual(x.Extends, y.Extends) &&
		NodeEqual(x.Implements, y.Implements) &&
		NodeSliceEqual(x.Stmts, y.Stmts)
}

func classClone(x ir.Class) ir.Class {
	return ir.Class{
		PhpDocComment: x.PhpDocComment,
		Extends:       NodeClone(x.Extends).(*ir.ClassExtendsStmt),
		Implements:    NodeClone(x.Implements).(*ir.ClassImplementsStmt),
		Stmts:         NodeSliceClone(x.Stmts),
	}
}
