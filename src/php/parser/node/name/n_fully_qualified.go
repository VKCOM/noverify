package name

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// FullyQualified node
type FullyQualified struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []node.Node
}

// NewFullyQualified node constructor
func NewFullyQualified(Parts []node.Node) *FullyQualified {
	return &FullyQualified{
		FreeFloating: nil,
		Parts:        Parts,
	}
}

// SetPosition sets node position
func (n *FullyQualified) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *FullyQualified) GetPosition() *position.Position {
	return n.Position
}

func (n *FullyQualified) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *FullyQualified) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
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
func (n *FullyQualified) GetParts() []node.Node {
	return n.Parts
}
