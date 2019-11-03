package node

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Root node
type Root struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []Node
}

// NewRoot node constructor
func NewRoot(Stmts []Node) *Root {
	return &Root{
		FreeFloating: nil,
		Stmts:        Stmts,
	}
}

// SetPosition sets node position
func (n *Root) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Root) GetPosition() *position.Position {
	return n.Position
}

func (n *Root) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Root) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Stmts != nil {
		for _, nn := range n.Stmts {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
