package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Ternary node
type Ternary struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Condition    node.Node
	IfTrue       node.Node
	IfFalse      node.Node
}

// NewTernary node constructor
func NewTernary(Condition node.Node, IfTrue node.Node, IfFalse node.Node) *Ternary {
	return &Ternary{
		FreeFloating: nil,
		Condition:    Condition,
		IfTrue:       IfTrue,
		IfFalse:      IfFalse,
	}
}

// SetPosition sets node position
func (n *Ternary) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Ternary) GetPosition() *position.Position {
	return n.Position
}

func (n *Ternary) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Ternary) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Condition != nil {
		n.Condition.Walk(v)
	}

	if n.IfTrue != nil {
		n.IfTrue.Walk(v)
	}

	if n.IfFalse != nil {
		n.IfFalse.Walk(v)
	}

	v.LeaveNode(n)
}
