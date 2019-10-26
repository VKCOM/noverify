package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ClassConstFetch node
type ClassConstFetch struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        node.Node
	ConstantName *node.Identifier
}

// NewClassConstFetch node constructor
func NewClassConstFetch(Class node.Node, ConstantName *node.Identifier) *ClassConstFetch {
	return &ClassConstFetch{
		FreeFloating: nil,
		Class:        Class,
		ConstantName: ConstantName,
	}
}

// SetPosition sets node position
func (n *ClassConstFetch) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ClassConstFetch) GetPosition() *position.Position {
	return n.Position
}

func (n *ClassConstFetch) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ClassConstFetch) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Class != nil {
		n.Class.Walk(v)
	}

	if n.ConstantName != nil {
		n.ConstantName.Walk(v)
	}

	v.LeaveNode(n)
}
