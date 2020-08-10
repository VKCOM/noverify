package irutil

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irfmt"
)

//go:generate go run ./codegen.go

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
		*ir.AssignMul:
		return true
	default:
		return false
	}
}

// FmtNode returns string representation of n.
func FmtNode(n ir.Node) string {
	return irfmt.Node(n)
}
