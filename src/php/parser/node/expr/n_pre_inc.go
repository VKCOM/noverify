package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// PreInc node
type PreInc struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
}

// NewPreInc node constructor
func NewPreInc(Variable node.Node) *PreInc {
	return &PreInc{
		FreeFloating: nil,
		Variable:     Variable,
	}
}

// SetPosition sets node position
func (n *PreInc) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *PreInc) GetPosition() *position.Position {
	return n.Position
}

func (n *PreInc) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *PreInc) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
	}

	v.LeaveNode(n)
}
