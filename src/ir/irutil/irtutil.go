package irutil

import (
	"github.com/VKCOM/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irfmt"
	"github.com/VKCOM/noverify/src/phpdoc"
)

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

func InLoop(path NodePath) bool {
	for i := 0; path.NthParent(i) != nil; i++ {
		cur := path.NthParent(i)
		if IsLoop(cur) {
			return true
		}
	}
	return false
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

func IsLoop(n ir.Node) bool {
	switch n.(type) {
	case *ir.ForStmt,
		*ir.ForeachStmt,
		*ir.WhileStmt,
		*ir.DoStmt:
		return true
	default:
		return false
	}
}

func IsBoolAnd(n ir.Node) bool {
	_, ok := n.(*ir.BooleanAndExpr)
	return ok
}

func IsBoolOr(n ir.Node) bool {
	_, ok := n.(*ir.BooleanOrExpr)
	return ok
}

// FmtNode returns string representation of n.
func FmtNode(n ir.Node) string {
	return irfmt.Node(n)
}

// FindChild searches for a child node in the passed
// subtree using a predicate.
//
// If the predicate returns true, the search ends.
func FindChild(where ir.Node, cb func(ir.Node) bool) ir.Node {
	if where == nil {
		return nil
	}
	w := newFindWalkerWithPredicate(where, cb)
	where.Walk(w)
	return w.res
}

// FindParent searches for a parent node for the passed
// node using a predicate.
//
// If the predicate returns true, the search ends.
func FindParent(n ir.Node, cb func(ir.Node) bool) ir.Node {
	if n == nil {
		return nil
	}

	for n != nil {
		if cb(n) {
			return n
		}
		n = n.Parent()
	}

	return nil
}

// FindPHPDoc searches for phpdoc by traversing all subtree and all tokens.
func FindPHPDoc(n ir.Node, withSuspicious bool) (doc string, found bool) {
	Inspect(n, func(n ir.Node) (continueTraverse bool) {
		n.IterateTokens(func(t *token.Token) (continueTraverse bool) {
			if t.ID == token.T_DOC_COMMENT {
				doc = string(t.Value)
				return false
			}

			if withSuspicious && t.ID == token.T_COMMENT && phpdoc.IsSuspicious(t.Value) {
				doc = string(t.Value)
				return false
			}

			return true
		})

		return doc == ""
	})

	if doc != "" {
		return doc, true
	}

	return doc, false
}

func classEqual(x, y ir.Class) bool {
	return x.Doc.Raw == y.Doc.Raw &&
		NodeEqual(x.Extends, y.Extends) &&
		NodeEqual(x.Implements, y.Implements) &&
		NodeSliceEqual(x.Stmts, y.Stmts)
}

func classClone(x ir.Class) ir.Class {
	return ir.Class{
		Doc: phpdoc.Comment{
			Raw: x.Doc.Raw,
		},
		Extends:    NodeClone(x.Extends).(*ir.ClassExtendsStmt),
		Implements: NodeClone(x.Implements).(*ir.ClassImplementsStmt),
		Stmts:      NodeSliceClone(x.Stmts),
	}
}

func FindClassMethodNode(n ir.Node, name string) *ir.ClassMethodStmt {
	class, ok := n.(*ir.ClassStmt)
	if !ok {
		return nil
	}

	for _, stmt := range class.Stmts {
		method, ok := stmt.(*ir.ClassMethodStmt)
		if !ok {
			continue
		}

		if method.MethodName.Value == name {
			return method
		}
	}

	return nil
}
