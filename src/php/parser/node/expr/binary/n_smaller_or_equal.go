package binary

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// SmallerOrEqual node
type SmallerOrEqual struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         node.Node
	Right        node.Node
}

// NewSmallerOrEqual node constructor
func NewSmallerOrEqual(Variable node.Node, Expression node.Node) *SmallerOrEqual {
	return &SmallerOrEqual{
		FreeFloating: nil,
		Left:         Variable,
		Right:        Expression,
	}
}

// SetPosition sets node position
func (n *SmallerOrEqual) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *SmallerOrEqual) GetPosition() *position.Position {
	return n.Position
}

func (n *SmallerOrEqual) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *SmallerOrEqual) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Left != nil {
		n.Left.Walk(v)
	}

	if n.Right != nil {
		n.Right.Walk(v)
	}

	v.LeaveNode(n)
}
