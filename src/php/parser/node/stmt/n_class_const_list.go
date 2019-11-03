package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ClassConstList node
type ClassConstList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Modifiers    []*node.Identifier
	Consts       []node.Node
}

// NewClassConstList node constructor
func NewClassConstList(Modifiers []*node.Identifier, Consts []node.Node) *ClassConstList {
	return &ClassConstList{
		FreeFloating: nil,
		Modifiers:    Modifiers,
		Consts:       Consts,
	}
}

// SetPosition sets node position
func (n *ClassConstList) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ClassConstList) GetPosition() *position.Position {
	return n.Position
}

func (n *ClassConstList) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ClassConstList) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Modifiers != nil {
		for _, nn := range n.Modifiers {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.Consts != nil {
		for _, nn := range n.Consts {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
