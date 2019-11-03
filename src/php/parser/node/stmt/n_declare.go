package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Declare node
type Declare struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Consts       []node.Node
	Stmt         node.Node
	Alt          bool
}

// NewDeclare node constructor
func NewDeclare(Consts []node.Node, Stmt node.Node, alt bool) *Declare {
	return &Declare{
		FreeFloating: nil,
		Consts:       Consts,
		Stmt:         Stmt,
		Alt:          alt,
	}
}

// SetPosition sets node position
func (n *Declare) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Declare) GetPosition() *position.Position {
	return n.Position
}

func (n *Declare) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Declare) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Consts != nil {
		for _, nn := range n.Consts {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}

	v.LeaveNode(n)
}
