package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Use node
type Use struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	UseType      *node.Identifier
	Use          node.Node
	Alias        *node.Identifier
}

// NewUse node constructor
func NewUse(use node.Node, Alias *node.Identifier) *Use {
	return &Use{
		FreeFloating: nil,
		Use:          use,
		Alias:        Alias,
	}
}

// SetPosition sets node position
func (n *Use) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Use) GetPosition() *position.Position {
	return n.Position
}

func (n *Use) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// SetUseType set use type and returns node
func (n *Use) SetUseType(UseType *node.Identifier) node.Node {
	n.UseType = UseType
	return n
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Use) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.UseType != nil {
		n.UseType.Walk(v)
	}

	if n.Use != nil {
		n.Use.Walk(v)
	}

	if n.Alias != nil {
		n.Alias.Walk(v)
	}

	v.LeaveNode(n)
}
