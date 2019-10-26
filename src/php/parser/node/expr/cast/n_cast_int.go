package cast

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Int node
type Int struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewInt node constructor
func NewInt(Expr node.Node) *Int {
	return &Int{
		FreeFloating: nil,
		Expr:         Expr,
	}
}

// SetPosition sets node position
func (n *Int) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Int) GetPosition() *position.Position {
	return n.Position
}

func (n *Int) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Int) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
