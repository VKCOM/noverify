package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Label node
type Label struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	LabelName    *node.Identifier
}

// NewLabel node constructor
func NewLabel(LabelName *node.Identifier) *Label {
	return &Label{
		FreeFloating: nil,
		LabelName:    LabelName,
	}
}

// SetPosition sets node position
func (n *Label) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Label) GetPosition() *position.Position {
	return n.Position
}

func (n *Label) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Label) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.LabelName != nil {
		n.LabelName.Walk(v)
	}

	v.LeaveNode(n)
}
