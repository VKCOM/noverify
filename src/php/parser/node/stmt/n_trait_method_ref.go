package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// TraitMethodRef node
type TraitMethodRef struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Trait        node.Node
	Method       *node.Identifier
}

// NewTraitMethodRef node constructor
func NewTraitMethodRef(Trait node.Node, Method *node.Identifier) *TraitMethodRef {
	return &TraitMethodRef{
		FreeFloating: nil,
		Trait:        Trait,
		Method:       Method,
	}
}

// SetPosition sets node position
func (n *TraitMethodRef) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *TraitMethodRef) GetPosition() *position.Position {
	return n.Position
}

func (n *TraitMethodRef) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *TraitMethodRef) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Trait != nil {
		n.Trait.Walk(v)
	}

	if n.Method != nil {
		n.Method.Walk(v)
	}

	v.LeaveNode(n)
}
