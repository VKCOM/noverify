package binary

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// NotIdentical node
type NotIdentical struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         node.Node
	Right        node.Node
}

// NewNotIdentical node constructor
func NewNotIdentical(Variable node.Node, Expression node.Node) *NotIdentical {
	return &NotIdentical{
		FreeFloating: nil,
		Left:         Variable,
		Right:        Expression,
	}
}

// SetPosition sets node position
func (n *NotIdentical) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *NotIdentical) GetPosition() *position.Position {
	return n.Position
}

func (n *NotIdentical) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *NotIdentical) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Left != nil {
		n.Left.Walk(v)
	}

	if n.Right != nil {
		n.Right.Walk(v)
	}

	v.LeaveNode(n)
}
