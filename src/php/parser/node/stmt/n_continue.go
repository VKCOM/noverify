package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Continue node
type Continue struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewContinue node constructor
func NewContinue(Expr node.Node) *Continue {
	return &Continue{
		FreeFloating: nil,
		Expr:         Expr,
	}
}

// SetPosition sets node position
func (n *Continue) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Continue) GetPosition() *position.Position {
	return n.Position
}

func (n *Continue) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Continue) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
