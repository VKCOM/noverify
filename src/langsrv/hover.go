package langsrv

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/state"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/position"
	"github.com/z7zmey/php-parser/walker"
)

type hoverWalker struct {
	position  int
	positions position.Positions
	n         node.Node
	st        meta.ClassParseState
}

// EnterNode is invoked at every node in hierarchy
func (d *hoverWalker) EnterNode(w walker.Walkable) bool {
	state.EnterNode(&d.st, w)
	return true
}

// GetChildrenVisitor is invoked at every node parameter that contains children nodes
func (d *hoverWalker) GetChildrenVisitor(key string) walker.Visitor {
	return d
}

// LeaveNode is invoked after node process
func (d *hoverWalker) LeaveNode(w walker.Walkable) {
	if d.n != nil {
		return
	}

	checkPos := false

	n := w.(node.Node)
	switch n.(type) {
	case *expr.Variable, *expr.MethodCall, *expr.FunctionCall, *expr.StaticCall:
		checkPos = true
	}

	state.LeaveNode(&d.st, w)

	if checkPos {
		pos := d.positions[n]

		if d.position > pos.EndPos || d.position < pos.StartPos {
			return
		}

		d.n = n
	}
}
