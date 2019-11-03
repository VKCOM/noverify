package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// TraitUse node
type TraitUse struct {
	FreeFloating        freefloating.Collection
	Position            *position.Position
	Traits              []node.Node
	TraitAdaptationList node.Node
}

// NewTraitUse node constructor
func NewTraitUse(Traits []node.Node, InnerAdaptationList node.Node) *TraitUse {
	return &TraitUse{
		FreeFloating:        nil,
		Traits:              Traits,
		TraitAdaptationList: InnerAdaptationList,
	}
}

// SetPosition sets node position
func (n *TraitUse) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *TraitUse) GetPosition() *position.Position {
	return n.Position
}

func (n *TraitUse) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *TraitUse) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Traits != nil {
		for _, nn := range n.Traits {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.TraitAdaptationList != nil {
		n.TraitAdaptationList.Walk(v)
	}

	v.LeaveNode(n)
}
