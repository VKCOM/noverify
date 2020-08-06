package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ArrayItem node
type ArrayItem struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Key          node.Node
	Val          node.Node
	Unpack       bool
}

// NewArrayItem node constructor
func NewArrayItem(Key node.Node, Val node.Node, Unpack bool) *ArrayItem {
	return &ArrayItem{
		FreeFloating: nil,
		Key:          Key,
		Val:          Val,
		Unpack:       Unpack,
	}
}

// SetPosition sets node position
func (n *ArrayItem) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ArrayItem) GetPosition() *position.Position {
	return n.Position
}

func (n *ArrayItem) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ArrayItem) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Key != nil {
		n.Key.Walk(v)
	}

	if n.Val != nil {
		n.Val.Walk(v)
	}

	v.LeaveNode(n)
}
