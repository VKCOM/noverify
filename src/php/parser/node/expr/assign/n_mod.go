package assign

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Mod node
type Mod struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
	Expression   node.Node
}

// NewMod node constructor
func NewMod(Variable node.Node, Expression node.Node) *Mod {
	return &Mod{
		FreeFloating: nil,
		Variable:     Variable,
		Expression:   Expression,
	}
}

// SetPosition sets node position
func (n *Mod) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Mod) GetPosition() *position.Position {
	return n.Position
}

func (n *Mod) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Mod) Walk(v walker.Visitor) {
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
