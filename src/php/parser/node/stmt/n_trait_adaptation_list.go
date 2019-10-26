package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// TraitAdaptationList node
type TraitAdaptationList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Adaptations  []node.Node
}

// NewTraitAdaptationList node constructor
func NewTraitAdaptationList(Adaptations []node.Node) *TraitAdaptationList {
	return &TraitAdaptationList{
		FreeFloating: nil,
		Adaptations:  Adaptations,
	}
}

// SetPosition sets node position
func (n *TraitAdaptationList) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *TraitAdaptationList) GetPosition() *position.Position {
	return n.Position
}

func (n *TraitAdaptationList) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *TraitAdaptationList) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Adaptations != nil {
		for _, nn := range n.Adaptations {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
