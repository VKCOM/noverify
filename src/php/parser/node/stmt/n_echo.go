package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Echo node
type Echo struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Exprs        []node.Node
}

// NewEcho node constructor
func NewEcho(Exprs []node.Node) *Echo {
	return &Echo{
		FreeFloating: nil,
		Exprs:        Exprs,
	}
}

// SetPosition sets node position
func (n *Echo) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Echo) GetPosition() *position.Position {
	return n.Position
}

func (n *Echo) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Echo) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Exprs != nil {
		for _, nn := range n.Exprs {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
