package assign

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Minus node
type Minus struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
	Expression   node.Node
}

// NewMinus node constructor
func NewMinus(Variable node.Node, Expression node.Node) *Minus {
	return &Minus{
		FreeFloating: nil,
		Variable:     Variable,
		Expression:   Expression,
	}
}

// SetPosition sets node position
func (n *Minus) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Minus) GetPosition() *position.Position {
	return n.Position
}

func (n *Minus) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Minus) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
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
