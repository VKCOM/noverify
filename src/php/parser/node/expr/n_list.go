package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// List node
type List struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Items        []*ArrayItem
	ShortSyntax  bool // Whether [] syntax is used instead of list() syntax
}

// NewList node constructor
func NewList(Items []*ArrayItem) *List {
	return &List{
		FreeFloating: nil,
		Items:        Items,
	}
}

// SetPosition sets node position
func (n *List) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *List) GetPosition() *position.Position {
	return n.Position
}

func (n *List) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Attributes returns node attributes as map
func (n *List) Attributes() map[string]interface{} {
	return nil
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *List) Walk(v walker.Visitor) {
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
