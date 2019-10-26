package node

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Identifier node
type Identifier struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// NewIdentifier node constructor
func NewIdentifier(Value string) *Identifier {
	return &Identifier{
		FreeFloating: nil,
		Value:        Value,
	}
}

// SetPosition sets node position
func (n *Identifier) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Identifier) GetPosition() *position.Position {
	return n.Position
}

func (n *Identifier) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Identifier) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	v.LeaveNode(n)
}
