package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ArrayDimFetch node
type ArrayDimFetch struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
	Dim          node.Node
}

// NewArrayDimFetch node constructor
func NewArrayDimFetch(Variable node.Node, Dim node.Node) *ArrayDimFetch {
	return &ArrayDimFetch{
		FreeFloating: nil,
		Variable:     Variable,
		Dim:          Dim,
	}
}

// SetPosition sets node position
func (n *ArrayDimFetch) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ArrayDimFetch) GetPosition() *position.Position {
	return n.Position
}

func (n *ArrayDimFetch) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ArrayDimFetch) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
	}

	if n.Dim != nil {
		n.Dim.Walk(v)
	}

	v.LeaveNode(n)
}
