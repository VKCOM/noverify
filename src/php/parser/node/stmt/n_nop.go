package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Nop node
type Nop struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
}

// NewNop node constructor
func NewNop() *Nop {
	return &Nop{}
}

// SetPosition sets node position
func (n *Nop) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Nop) GetPosition() *position.Position {
	return n.Position
}

func (n *Nop) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Nop) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	v.LeaveNode(n)
}
