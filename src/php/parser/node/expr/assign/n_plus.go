package assign

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Plus node
type Plus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
	Expression   node.Node
}

// NewPlus node constructor
func NewPlus(Variable node.Node, Expression node.Node) *Plus {
	return &Plus{
		FreeFloating: nil,
		Variable:     Variable,
		Expression:   Expression,
	}
}

// SetPosition sets node position
func (n *Plus) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Plus) GetPosition() *position.Position {
	return n.Position
}

func (n *Plus) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Plus) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
	}

	if n.Expression != nil {
		n.Expression.Walk(v)
	}

	v.LeaveNode(n)
}
