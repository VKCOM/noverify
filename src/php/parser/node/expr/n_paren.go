package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Paren is a parenthesized expression.
type Paren struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node // Parenthesized expression
}

// SetPosition sets node position
func (n *Paren) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Paren) GetPosition() *position.Position {
	return n.Position
}

func (n *Paren) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Paren) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
