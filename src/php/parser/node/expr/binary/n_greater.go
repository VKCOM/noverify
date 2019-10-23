package binary

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Greater node
type Greater struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Left         node.Node
	Right        node.Node
}

// NewGreater node constructor
func NewGreater(Variable node.Node, Expression node.Node) *Greater {
	return &Greater{
		FreeFloating: nil,
		Left:         Variable,
		Right:        Expression,
	}
}

// SetPosition sets node position
func (n *Greater) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Greater) GetPosition() *position.Position {
	return n.Position
}

func (n *Greater) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Attributes returns node attributes as map
func (n *Greater) Attributes() map[string]interface{} {
	return nil
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Greater) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Left != nil {
		v.EnterChildNode("Left", n)
		n.Left.Walk(v)
		v.LeaveChildNode("Left", n)
	}

	if n.Right != nil {
		v.EnterChildNode("Right", n)
		n.Right.Walk(v)
		v.LeaveChildNode("Right", n)
	}

	v.LeaveNode(n)
}
