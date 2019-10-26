package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Property node
type Property struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	Variable      *node.SimpleVar
	Expr          node.Node
}

// NewProperty node constructor
func NewProperty(Variable *node.SimpleVar, Expr node.Node, PhpDocComment string) *Property {
	return &Property{
		FreeFloating:  nil,
		PhpDocComment: PhpDocComment,
		Variable:      Variable,
		Expr:          Expr,
	}
}

// SetPosition sets node position
func (n *Property) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Property) GetPosition() *position.Position {
	return n.Position
}

func (n *Property) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Property) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
