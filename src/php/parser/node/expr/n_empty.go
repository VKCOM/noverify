package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Empty node
type Empty struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewEmpty node constructor
func NewEmpty(Expression node.Node) *Empty {
	return &Empty{
		FreeFloating: nil,
		Expr:         Expression,
	}
}

// SetPosition sets node position
func (n *Empty) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Empty) GetPosition() *position.Position {
	return n.Position
}

func (n *Empty) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Empty) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
