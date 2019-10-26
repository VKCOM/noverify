package langsrv

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/state"
)

type hoverWalker struct {
	position int
	n        node.Node
	st       meta.ClassParseState
}

// EnterNode is invoked at every node in hierarchy
func (d *hoverWalker) EnterNode(w walker.Walkable) bool {
	state.EnterNode(&d.st, w)
	return true
}

// LeaveNode is invoked after node process
func (d *hoverWalker) LeaveNode(w walker.Walkable) {
	if d.n != nil {
		return
	}

	checkPos := false

	n := w.(node.Node)
	switch n.(type) {
	case *node.Variable, *expr.MethodCall, *expr.FunctionCall, *expr.StaticCall:
		checkPos = true
	}

	state.LeaveNode(&d.st, w)

	if checkPos {
		pos := n.GetPosition()

		if d.position > pos.EndPos || d.position < pos.StartPos {
			return
		}

		d.n = n
	}
}
