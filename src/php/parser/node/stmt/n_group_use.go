package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// GroupUse node
type GroupUse struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	UseType      node.Node
	Prefix       node.Node
	UseList      []node.Node
}

// NewGroupUse node constructor
func NewGroupUse(UseType node.Node, Prefix node.Node, UseList []node.Node) *GroupUse {
	return &GroupUse{
		FreeFloating: nil,
		UseType:      UseType,
		Prefix:       Prefix,
		UseList:      UseList,
	}
}

// SetPosition sets node position
func (n *GroupUse) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *GroupUse) GetPosition() *position.Position {
	return n.Position
}

func (n *GroupUse) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// SetUseType set use type and returns node
func (n *GroupUse) SetUseType(UseType node.Node) node.Node {
	n.UseType = UseType
	return n
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *GroupUse) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.UseType != nil {
		n.UseType.Walk(v)
	}

	if n.Prefix != nil {
		n.Prefix.Walk(v)
	}

	if n.UseList != nil {
		for _, nn := range n.UseList {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
