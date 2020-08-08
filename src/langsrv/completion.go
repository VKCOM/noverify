package langsrv

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/state"
)

type completionWalker struct {
	// params
	position int
	scopes   map[ir.Node]*meta.Scope

	// output
	foundScope *meta.Scope
	st         meta.ClassParseState
}

// EnterNode is invoked at every node in hierarchy
func (d *completionWalker) EnterNode(w ir.Node) bool {
	state.EnterNode(&d.st, w)

	return d.foundScope == nil
}

// LeaveNode is invoked after node process
func (d *completionWalker) LeaveNode(n ir.Node) {
	if d.foundScope != nil {
		return
	}

	state.LeaveNode(&d.st, n)

	pos := ir.GetPosition(n)

	if pos == nil {
		return
	}

	if d.position > pos.EndPos || d.position < pos.StartPos {
		return
	}

	sc, ok := d.scopes[n]
	if !ok {
		return
	}

	d.foundScope = sc
}
