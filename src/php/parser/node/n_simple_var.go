package node

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// SimpleVar node is a plain PHP variable where it's name is a simple identifier.
//
// For example, $x is a simple var.
// $$x is not a simple var, but rather a Variable, where VarName is a SimpleVar.
type SimpleVar struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Name         string // Variable name without leading "$"
}

// NewSimpleVar node constructor
func NewSimpleVar(name string) *SimpleVar {
	return &SimpleVar{
		FreeFloating: nil,
		Name:         name,
	}
}

// SetPosition sets node position
func (n *SimpleVar) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *SimpleVar) GetPosition() *position.Position {
	return n.Position
}

func (n *SimpleVar) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *SimpleVar) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	v.LeaveNode(n)
}
