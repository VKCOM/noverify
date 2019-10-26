package name

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Name node
type Name struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []node.Node
}

// NewName node constructor
func NewName(Parts []node.Node) *Name {
	return &Name{
		FreeFloating: nil,
		Parts:        Parts,
	}
}

// SetPosition sets node position
func (n *Name) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Name) GetPosition() *position.Position {
	return n.Position
}

func (n *Name) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Name) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Parts != nil {
		for _, nn := range n.Parts {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}

// GetParts returns the name parts
func (n *Name) GetParts() []node.Node {
	return n.Parts
}
