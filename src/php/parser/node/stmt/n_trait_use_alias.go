package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// TraitUseAlias node
type TraitUseAlias struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Ref          node.Node
	Modifier     node.Node
	Alias        *node.Identifier
}

// NewTraitUseAlias node constructor
func NewTraitUseAlias(Ref node.Node, Modifier node.Node, Alias *node.Identifier) *TraitUseAlias {
	return &TraitUseAlias{
		FreeFloating: nil,
		Ref:          Ref,
		Modifier:     Modifier,
		Alias:        Alias,
	}
}

// SetPosition sets node position
func (n *TraitUseAlias) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *TraitUseAlias) GetPosition() *position.Position {
	return n.Position
}

func (n *TraitUseAlias) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *TraitUseAlias) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Ref != nil {
		n.Ref.Walk(v)
	}

	if n.Modifier != nil {
		n.Modifier.Walk(v)
	}

	if n.Alias != nil {
		n.Alias.Walk(v)
	}

	v.LeaveNode(n)
}
