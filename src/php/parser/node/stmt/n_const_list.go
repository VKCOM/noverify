package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ConstList node
type ConstList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Consts       []node.Node
}

// NewConstList node constructor
func NewConstList(Consts []node.Node) *ConstList {
	return &ConstList{
		FreeFloating: nil,
		Consts:       Consts,
	}
}

// SetPosition sets node position
func (n *ConstList) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ConstList) GetPosition() *position.Position {
	return n.Position
}

func (n *ConstList) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ConstList) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
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
