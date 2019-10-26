package assign

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Assign node
type Assign struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
	Expression   node.Node
}

// NewAssign node constructor
func NewAssign(Variable node.Node, Expression node.Node) *Assign {
	return &Assign{
		FreeFloating: nil,
		Variable:     Variable,
		Expression:   Expression,
	}
}

// SetPosition sets node position
func (n *Assign) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Assign) GetPosition() *position.Position {
	return n.Position
}

func (n *Assign) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Assign) Walk(v walker.Visitor) {
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
