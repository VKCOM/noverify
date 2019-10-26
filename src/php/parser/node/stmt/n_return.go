package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Return node
type Return struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewReturn node constructor
func NewReturn(Expr node.Node) *Return {
	return &Return{
		FreeFloating: nil,
		Expr:         Expr,
	}
}

// SetPosition sets node position
func (n *Return) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Return) GetPosition() *position.Position {
	return n.Position
}

func (n *Return) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Return) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
