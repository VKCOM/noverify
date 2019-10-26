package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Yield node
type Yield struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Key          node.Node
	Value        node.Node
}

// NewYield node constructor
func NewYield(Key node.Node, Value node.Node) *Yield {
	return &Yield{
		FreeFloating: nil,
		Key:          Key,
		Value:        Value,
	}
}

// SetPosition sets node position
func (n *Yield) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Yield) GetPosition() *position.Position {
	return n.Position
}

func (n *Yield) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Yield) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Key != nil {
		n.Key.Walk(v)
	}

	if n.Value != nil {
		n.Value.Walk(v)
	}

	v.LeaveNode(n)
}
