package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// PreDec node
type PreDec struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
}

// NewPreDec node constructor
func NewPreDec(Variable node.Node) *PreDec {
	return &PreDec{
		FreeFloating: nil,
		Variable:     Variable,
	}
}

// SetPosition sets node position
func (n *PreDec) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *PreDec) GetPosition() *position.Position {
	return n.Position
}

func (n *PreDec) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *PreDec) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
	}

	v.LeaveNode(n)
}
