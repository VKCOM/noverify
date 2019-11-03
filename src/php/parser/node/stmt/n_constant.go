package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Constant node
type Constant struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	ConstantName  *node.Identifier
	Expr          node.Node
}

// NewConstant node constructor
func NewConstant(ConstantName *node.Identifier, Expr node.Node, PhpDocComment string) *Constant {
	return &Constant{
		FreeFloating:  nil,
		PhpDocComment: PhpDocComment,
		ConstantName:  ConstantName,
		Expr:          Expr,
	}
}

// SetPosition sets node position
func (n *Constant) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Constant) GetPosition() *position.Position {
	return n.Position
}

func (n *Constant) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Constant) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.ConstantName != nil {
		n.ConstantName.Walk(v)
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
