package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// InstanceOf node
type InstanceOf struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
	Class        node.Node
}

// NewInstanceOf node constructor
func NewInstanceOf(Expr node.Node, Class node.Node) *InstanceOf {
	return &InstanceOf{
		FreeFloating: nil,
		Expr:         Expr,
		Class:        Class,
	}
}

// SetPosition sets node position
func (n *InstanceOf) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *InstanceOf) GetPosition() *position.Position {
	return n.Position
}

func (n *InstanceOf) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *InstanceOf) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	if n.Class != nil {
		n.Class.Walk(v)
	}

	v.LeaveNode(n)
}
