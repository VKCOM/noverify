package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Goto node
type Goto struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Label        *node.Identifier
}

// NewGoto node constructor
func NewGoto(Label *node.Identifier) *Goto {
	return &Goto{
		FreeFloating: nil,
		Label:        Label,
	}
}

// SetPosition sets node position
func (n *Goto) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Goto) GetPosition() *position.Position {
	return n.Position
}

func (n *Goto) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Goto) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Label != nil {
		n.Label.Walk(v)
	}

	v.LeaveNode(n)
}
