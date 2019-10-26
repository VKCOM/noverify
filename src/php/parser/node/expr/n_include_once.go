package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// IncludeOnce node
type IncludeOnce struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewIncludeOnce node constructor
func NewIncludeOnce(Expression node.Node) *IncludeOnce {
	return &IncludeOnce{
		FreeFloating: nil,
		Expr:         Expression,
	}
}

// SetPosition sets node position
func (n *IncludeOnce) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *IncludeOnce) GetPosition() *position.Position {
	return n.Position
}

func (n *IncludeOnce) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *IncludeOnce) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
