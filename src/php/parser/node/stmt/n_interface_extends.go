package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// InterfaceExtends node
type InterfaceExtends struct {
	FreeFloating   freefloating.Collection
	Position       *position.Position
	InterfaceNames []node.Node
}

// NewInterfaceExtends node constructor
func NewInterfaceExtends(InterfaceNames []node.Node) *InterfaceExtends {
	return &InterfaceExtends{
		FreeFloating:   nil,
		InterfaceNames: InterfaceNames,
	}
}

// SetPosition sets node position
func (n *InterfaceExtends) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *InterfaceExtends) GetPosition() *position.Position {
	return n.Position
}

func (n *InterfaceExtends) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *InterfaceExtends) Walk(v walker.Visitor) {
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
