package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Case node
type Case struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         node.Node
	Stmts        []node.Node
}

// NewCase node constructor
func NewCase(Cond node.Node, Stmts []node.Node) *Case {
	return &Case{
		FreeFloating: nil,
		Cond:         Cond,
		Stmts:        Stmts,
	}
}

// SetPosition sets node position
func (n *Case) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Case) GetPosition() *position.Position {
	return n.Position
}

func (n *Case) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Case) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Cond != nil {
		n.Cond.Walk(v)
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
