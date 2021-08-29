package irutil

import (
	"github.com/VKCOM/noverify/src/ir"
)

func Inspect(root ir.Node, visit func(ir.Node) bool) {
	if root == nil {
		return
	}
	w := inspectWalker{visit: visit}
	root.Walk(w)
}

type inspectWalker struct {
	visit func(n ir.Node) bool
}

func (w inspectWalker) EnterNode(n ir.Node) bool {
	return w.visit(n)
}

func (w inspectWalker) LeaveNode(n ir.Node) {}
