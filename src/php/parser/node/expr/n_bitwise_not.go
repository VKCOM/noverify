package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// BitwiseNot node
type BitwiseNot struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewBitwiseNot node constructor
func NewBitwiseNot(Expression node.Node) *BitwiseNot {
	return &BitwiseNot{
		FreeFloating: nil,
		Expr:         Expression,
	}
}

// SetPosition sets node position
func (n *BitwiseNot) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *BitwiseNot) GetPosition() *position.Position {
	return n.Position
}

func (n *BitwiseNot) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *BitwiseNot) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
