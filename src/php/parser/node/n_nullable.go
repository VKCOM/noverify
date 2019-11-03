package node

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Nullable node
type Nullable struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// NewNullable node constructor
func NewNullable(Expression Node) *Nullable {
	return &Nullable{
		FreeFloating: nil,
		Expr:         Expression,
	}
}

// SetPosition sets node position
func (n *Nullable) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Nullable) GetPosition() *position.Position {
	return n.Position
}

func (n *Nullable) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Nullable) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
