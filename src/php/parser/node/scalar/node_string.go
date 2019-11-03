package scalar

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// String node
type String struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Value        string
}

// NewString node constructor
func NewString(Value string) *String {
	return &String{
		FreeFloating: nil,
		Value:        Value,
	}
}

// SetPosition sets node position
func (n *String) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *String) GetPosition() *position.Position {
	return n.Position
}

func (n *String) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *String) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	v.LeaveNode(n)
}
