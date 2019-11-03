package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Do node
type Do struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmt         node.Node
	Cond         node.Node
}

// NewDo node constructor
func NewDo(Stmt node.Node, Cond node.Node) *Do {
	return &Do{
		FreeFloating: nil,
		Stmt:         Stmt,
		Cond:         Cond,
	}
}

// SetPosition sets node position
func (n *Do) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Do) GetPosition() *position.Position {
	return n.Position
}

func (n *Do) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Do) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}

	if n.Cond != nil {
		n.Cond.Walk(v)
	}

	v.LeaveNode(n)
}
