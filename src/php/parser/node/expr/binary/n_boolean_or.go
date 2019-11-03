package binary

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// BooleanOr node
type BooleanOr struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         node.Node
	Right        node.Node
}

// NewBooleanOr node constructor
func NewBooleanOr(Variable node.Node, Expression node.Node) *BooleanOr {
	return &BooleanOr{
		FreeFloating: nil,
		Left:         Variable,
		Right:        Expression,
	}
}

// SetPosition sets node position
func (n *BooleanOr) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *BooleanOr) GetPosition() *position.Position {
	return n.Position
}

func (n *BooleanOr) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *BooleanOr) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
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
