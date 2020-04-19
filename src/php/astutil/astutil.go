package astutil

import (
	"github.com/VKCOM/noverify/src/php/parser/node"
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
