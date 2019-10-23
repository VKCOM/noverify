package node

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Argument node
type Argument struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variadic     bool // if ... before variable
	IsReference  bool // if & before variable
	Expr         Node // Exression
}

// NewArgument node constructor
func NewArgument(Expression Node, Variadic bool, IsReference bool) *Argument {
	return &Argument{
		FreeFloating: nil,
		Variadic:     Variadic,
		IsReference:  IsReference,
		Expr:         Expression,
	}
}

// SetPosition sets node position
func (n *Argument) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Argument) GetPosition() *position.Position {
	return n.Position
}

func (n *Argument) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Attributes returns node attributes as map
func (n *Argument) Attributes() map[string]interface{} {
	return map[string]interface{}{
		"Variadic":    n.Variadic,
		"IsReference": n.IsReference,
	}
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Argument) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Expr != nil {
		v.EnterChildNode("Expr", n)
		n.Expr.Walk(v)
		v.LeaveChildNode("Expr", n)
	}

	v.LeaveNode(n)
}
