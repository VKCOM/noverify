package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ConstFetch node
type ConstFetch struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Constant     node.Node
}

// NewConstFetch node constructor
func NewConstFetch(Constant node.Node) *ConstFetch {
	return &ConstFetch{
		FreeFloating: nil,
		Constant:     Constant,
	}
}

// SetPosition sets node position
func (n *ConstFetch) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ConstFetch) GetPosition() *position.Position {
	return n.Position
}

func (n *ConstFetch) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ConstFetch) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Constant != nil {
		n.Constant.Walk(v)
	}

	v.LeaveNode(n)
}
