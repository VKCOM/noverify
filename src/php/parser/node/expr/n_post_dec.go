package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// PostDec node
type PostDec struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
}

// NewPostDec node constructor
func NewPostDec(Variable node.Node) *PostDec {
	return &PostDec{
		FreeFloating: nil,
		Variable:     Variable,
	}
}

// SetPosition sets node position
func (n *PostDec) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *PostDec) GetPosition() *position.Position {
	return n.Position
}

func (n *PostDec) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *PostDec) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
	}

	v.LeaveNode(n)
}
