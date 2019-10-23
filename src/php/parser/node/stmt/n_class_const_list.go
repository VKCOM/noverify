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

// Attributes returns node attributes as map
func (n *ClassConstList) Attributes() map[string]interface{} {
	return nil
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ClassConstList) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Modifiers != nil {
		v.EnterChildList("Modifiers", n)
		for _, nn := range n.Modifiers {
			if nn != nil {
				nn.Walk(v)
			}
		}
		v.LeaveChildList("Modifiers", n)
	}

	if n.Consts != nil {
		v.EnterChildList("Consts", n)
		for _, nn := range n.Consts {
			if nn != nil {
				nn.Walk(v)
			}
		}
		v.LeaveChildList("Consts", n)
	}

	v.LeaveNode(n)
}
