package langsrv

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/state"
)

type hoverWalker struct {
	position int
	n        ir.Node
	st       meta.ClassParseState
}

// EnterNode is invoked at every node in hierarchy
func (d *hoverWalker) EnterNode(n ir.Node) bool {
	state.EnterNode(&d.st, n)
	return true
}

// LeaveNode is invoked after node process
func (d *hoverWalker) LeaveNode(n ir.Node) {
	if d.n != nil {
		return
	}

	checkPos := false

	switch n.(type) {
	case *ir.SimpleVar, *ir.MethodCallExpr, *ir.FunctionCallExpr, *ir.StaticCallExpr:
		checkPos = true
	}

	state.LeaveNode(&d.st, n)

	if checkPos {
		pos := ir.GetPosition(n)

		if d.position > pos.EndPos || d.position < pos.StartPos {
			return
		}

		d.n = n
	}
}
