package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ClassImplements node
type ClassImplements struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	InterfaceNames []node.Node
}

// NewClassImplements node constructor
func NewClassImplements(interfaceNames []node.Node) *ClassImplements {
	return &ClassImplements{
		FreeFloating:   nil,
		InterfaceNames: interfaceNames,
	}
}

// SetPosition sets node position
func (n *ClassImplements) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ClassImplements) GetPosition() *position.Position {
	return n.Position
}

func (n *ClassImplements) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ClassImplements) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.InterfaceNames != nil {
		for _, nn := range n.InterfaceNames {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
