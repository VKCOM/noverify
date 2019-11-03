package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Require node
type Require struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewRequire node constructor
func NewRequire(Expression node.Node) *Require {
	return &Require{
		FreeFloating: nil,
		Expr:         Expression,
	}
}

// SetPosition sets node position
func (n *Require) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Require) GetPosition() *position.Position {
	return n.Position
}

func (n *Require) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Require) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
