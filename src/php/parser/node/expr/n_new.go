package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// New node
type New struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        node.Node
	ArgumentList *node.ArgumentList
}

// NewNew node constructor
func NewNew(Class node.Node, ArgumentList *node.ArgumentList) *New {
	return &New{
		FreeFloating: nil,
		Class:        Class,
		ArgumentList: ArgumentList,
	}
}

// SetPosition sets node position
func (n *New) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *New) GetPosition() *position.Position {
	return n.Position
}

func (n *New) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *New) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Class != nil {
		n.Class.Walk(v)
	}

	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}

	v.LeaveNode(n)
}
