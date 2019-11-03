package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Print node
type Print struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewPrint node constructor
func NewPrint(Expression node.Node) *Print {
	return &Print{
		FreeFloating: nil,
		Expr:         Expression,
	}
}

// SetPosition sets node position
func (n *Print) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Print) GetPosition() *position.Position {
	return n.Position
}

func (n *Print) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Print) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
