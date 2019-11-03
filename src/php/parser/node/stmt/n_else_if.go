package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ElseIf node
type ElseIf struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         node.Node
	Stmt         node.Node
	AltSyntax    bool // Whether alternative colon-style syntax is used
}

// NewElseIf node constructor
func NewElseIf(Cond node.Node, Stmt node.Node) *ElseIf {
	return &ElseIf{
		FreeFloating: nil,
		Cond:         Cond,
		Stmt:         Stmt,
	}
}

// SetPosition sets node position
func (n *ElseIf) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ElseIf) GetPosition() *position.Position {
	return n.Position
}

func (n *ElseIf) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ElseIf) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Cond != nil {
		n.Cond.Walk(v)
	}

	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}

	v.LeaveNode(n)
}
