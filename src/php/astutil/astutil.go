package astutil

import (
	"bytes"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/expr/assign"
	"github.com/VKCOM/noverify/src/php/parser/printer"
)

//go:generate go run ./gen_equal.go

func NodeSliceEqual(xs, ys []node.Node) bool {
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

func IsAssign(n node.Node) bool {
	switch n.(type) {
	case *assign.Assign,
		*assign.Concat,
		*assign.Plus,
		*assign.Reference,
		*assign.Div,
		*assign.Pow,
		*assign.BitwiseAnd,
		*assign.BitwiseOr,
		*assign.BitwiseXor,
		*assign.ShiftLeft,
		*assign.ShiftRight,
		*assign.Minus,
		*assign.Mod,
		*assign.Mul:
		return true
	default:
		return false
	}
}

// FmtNode is used for debug purposes and returns string representation of a specified node.
func FmtNode(n node.Node) string {
	var b bytes.Buffer
	printer.NewPrettyPrinter(&b, " ").Print(n)
	return b.String()
}

func ValidArrayKey(n node.Node) bool {
	switch n.(type) {
	case *expr.New, *expr.Closure, *expr.Array:
		return false
	}
	return true
}
