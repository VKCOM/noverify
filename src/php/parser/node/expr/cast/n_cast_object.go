package cast

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Object node
type Object struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
}

// NewObject node constructor
func NewObject(Expr node.Node) *Object {
	return &Object{
		FreeFloating: nil,
		Expr:         Expr,
	}
}

// SetPosition sets node position
func (n *Object) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Object) GetPosition() *position.Position {
	return n.Position
}

func (n *Object) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Object) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
