package langsrv

import (
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/state"
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/walker"
)

type completionWalker struct {
	// params
	position int
	scopes   map[node.Node]*meta.Scope

	// output
	foundScope *meta.Scope
	st         meta.ClassParseState
}

// EnterNode is invoked at every node in hierarchy
func (d *completionWalker) EnterNode(w walker.Walkable) bool {
	state.EnterNode(&d.st, w)

	return d.foundScope == nil
}

// GetChildrenVisitor is invoked at every node parameter that contains children nodes
func (d *completionWalker) GetChildrenVisitor(key string) walker.Visitor {
	return d
}

// LeaveNode is invoked after node process
func (d *completionWalker) LeaveNode(w walker.Walkable) {
	if d.foundScope != nil {
		return
	}

	state.LeaveNode(&d.st, w)

	n := w.(node.Node)
	pos := n.GetPosition()

	if pos == nil {
		return
	}

	if d.position > pos.EndPos || d.position < pos.StartPos {
		return
	}

	sc, ok := d.scopes[n.(node.Node)]
	if !ok {
		return
	}

	d.foundScope = sc
}

func (d *completionWalker) EnterChildNode(key string, w walker.Walkable) {}
func (d *completionWalker) LeaveChildNode(key string, w walker.Walkable) {}
func (d *completionWalker) EnterChildList(key string, w walker.Walkable) {}
func (d *completionWalker) LeaveChildList(key string, w walker.Walkable) {}
