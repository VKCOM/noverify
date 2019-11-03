package node

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Var is an expression that evaluates to a variable.
// Unlike SimpleVar, it can contain complex expressions.
//
// Here are some examples:
//	$x     - SimpleVar.Name="x"
//	$$x    - Var.Expr.(*SimpleVar).Name="x"
//	${"x"} - Var.Expr.(*scalar.String).Value="x"
type Var struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         Node
}

// NewVar node constructor
func NewVar(expr Node) *Var {
	return &Var{Expr: expr}
}

// SetPosition sets node position
func (n *Var) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Var) GetPosition() *position.Position {
	return n.Position
}

func (n *Var) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Var) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
