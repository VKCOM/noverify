package scalar

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// EncapsedStringPart node
type EncapsedStringPart struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// NewEncapsedStringPart node constructor
func NewEncapsedStringPart(Value string) *EncapsedStringPart {
	return &EncapsedStringPart{
		FreeFloating: nil,
		Value:        Value,
	}
}

// SetPosition sets node position
func (n *EncapsedStringPart) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *EncapsedStringPart) GetPosition() *position.Position {
	return n.Position
}

func (n *EncapsedStringPart) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *EncapsedStringPart) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	v.LeaveNode(n)
}
