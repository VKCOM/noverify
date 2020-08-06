package assign

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Coalesce node
type Coalesce struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
	Expression   node.Node
}

// NewCoalesce node constructor
func NewCoalesce(Variable node.Node, Expression node.Node) *Coalesce {
	return &Coalesce{
		FreeFloating: nil,
		Variable:     Variable,
		Expression:   Expression,
	}
}

// SetPosition sets node position
func (n *Coalesce) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Coalesce) GetPosition() *position.Position {
	return n.Position
}

func (n *Coalesce) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Coalesce) Walk(v walker.Visitor) {
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
