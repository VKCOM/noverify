package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Global node
type Global struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Vars         []node.Node
}

// NewGlobal node constructor
func NewGlobal(Vars []node.Node) *Global {
	return &Global{
		FreeFloating: nil,
		Vars:         Vars,
	}
}

// SetPosition sets node position
func (n *Global) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Global) GetPosition() *position.Position {
	return n.Position
}

func (n *Global) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Global) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Vars != nil {
		for _, nn := range n.Vars {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
