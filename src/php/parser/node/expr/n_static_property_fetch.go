package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// StaticPropertyFetch node
type StaticPropertyFetch struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Class        node.Node
	Property     node.Node
}

// NewStaticPropertyFetch node constructor
func NewStaticPropertyFetch(Class node.Node, Property node.Node) *StaticPropertyFetch {
	return &StaticPropertyFetch{
		FreeFloating: nil,
		Class:        Class,
		Property:     Property,
	}
}

// SetPosition sets node position
func (n *StaticPropertyFetch) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *StaticPropertyFetch) GetPosition() *position.Position {
	return n.Position
}

func (n *StaticPropertyFetch) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *StaticPropertyFetch) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Class != nil {
		n.Class.Walk(v)
	}

	if n.Property != nil {
		n.Property.Walk(v)
	}

	v.LeaveNode(n)
}
