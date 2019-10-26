package scalar

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// MagicConstant node
type MagicConstant struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// NewMagicConstant node constructor
func NewMagicConstant(Value string) *MagicConstant {
	return &MagicConstant{
		FreeFloating: nil,
		Value:        Value,
	}
}

// SetPosition sets node position
func (n *MagicConstant) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *MagicConstant) GetPosition() *position.Position {
	return n.Position
}

func (n *MagicConstant) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *MagicConstant) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	v.LeaveNode(n)
}
