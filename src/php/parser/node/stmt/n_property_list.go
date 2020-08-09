package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// PropertyList node
type PropertyList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Modifiers    []*node.Identifier
	Type         node.Node
	Properties   []node.Node
}

// NewPropertyList node constructor
func NewPropertyList(Modifiers []*node.Identifier, Type node.Node, Properties []node.Node) *PropertyList {
	return &PropertyList{
		FreeFloating: nil,
		Modifiers:    Modifiers,
		Type:         Type,
		Properties:   Properties,
	}
}

// SetPosition sets node position
func (n *PropertyList) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *PropertyList) GetPosition() *position.Position {
	return n.Position
}

func (n *PropertyList) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *PropertyList) Walk(v walker.Visitor) {
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

	if n.Type != nil {
		n.Type.Walk(v)
	}

	if n.Properties != nil {
		for _, nn := range n.Properties {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
