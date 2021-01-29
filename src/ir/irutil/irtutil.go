package irutil

import (
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

// Find searches for a node in the passed subtree.
func Find(what ir.Node, where ir.Node) bool {
	if what == nil || where == nil {
		return false
	}
	w := newFindWalker(what, where)
	where.Walk(w)
	return w.found
}

// FindWithPredicate searches for a node in the passed
// subtree using a predicate.
//
// If the predicate returns true, the search ends.
func FindWithPredicate(what ir.Node, where ir.Node, pred findPredicate) bool {
	if what == nil || where == nil {
		return false
	}
	w := newFindWalkerWithPredicate(what, where, pred)
	where.Walk(w)
	return w.found
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
