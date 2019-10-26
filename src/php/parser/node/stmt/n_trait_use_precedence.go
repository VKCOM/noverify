package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// TraitUsePrecedence node
type TraitUsePrecedence struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Ref          node.Node
	Insteadof    []node.Node
}

// NewTraitUsePrecedence node constructor
func NewTraitUsePrecedence(Ref node.Node, Insteadof []node.Node) *TraitUsePrecedence {
	return &TraitUsePrecedence{
		FreeFloating: nil,
		Ref:          Ref,
		Insteadof:    Insteadof,
	}
}

// SetPosition sets node position
func (n *TraitUsePrecedence) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *TraitUsePrecedence) GetPosition() *position.Position {
	return n.Position
}

func (n *TraitUsePrecedence) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *TraitUsePrecedence) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Ref != nil {
		n.Ref.Walk(v)
	}

	if n.Insteadof != nil {
		for _, nn := range n.Insteadof {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
