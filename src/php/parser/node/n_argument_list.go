package node

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ArgumentList node
type ArgumentList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Arguments    []Node
}

// NewArgumentList node constructor
func NewArgumentList(Arguments []Node) *ArgumentList {
	return &ArgumentList{
		FreeFloating: nil,
		Arguments:    Arguments,
	}
}

// SetPosition sets node position
func (n *ArgumentList) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ArgumentList) GetPosition() *position.Position {
	return n.Position
}

func (n *ArgumentList) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ArgumentList) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Arguments != nil {
		for _, nn := range n.Arguments {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
