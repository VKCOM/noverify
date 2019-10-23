package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Array node
type Array struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Items        []*ArrayItem
	ShortSyntax  bool // Whether [] syntax is used instead of array() syntax
}

// NewArray node constructor
func NewArray(Items []*ArrayItem) *Array {
	return &Array{
		FreeFloating: nil,
		Items:        Items,
	}
}

// SetPosition sets node position
func (n *Array) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Array) GetPosition() *position.Position {
	return n.Position
}

func (n *Array) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Attributes returns node attributes as map
func (n *Array) Attributes() map[string]interface{} {
	return nil
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Array) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Items != nil {
		v.EnterChildList("Items", n)
		for _, nn := range n.Items {
			if nn != nil {
				nn.Walk(v)
			}
		}
		v.LeaveChildList("Items", n)
	}

	v.LeaveNode(n)
}
