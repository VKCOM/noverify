package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Foreach node
type Foreach struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Expr         node.Node
	Key          node.Node
	Variable     node.Node
	Stmt         node.Node
	AltSyntax    bool // Whether alternative colon-style syntax is used
}

// NewForeach node constructor
func NewForeach(Expr node.Node, Key node.Node, Variable node.Node, Stmt node.Node) *Foreach {
	return &Foreach{
		FreeFloating: nil,
		Expr:         Expr,
		Key:          Key,
		Variable:     Variable,
		Stmt:         Stmt,
	}
}

// SetPosition sets node position
func (n *Foreach) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Foreach) GetPosition() *position.Position {
	return n.Position
}

func (n *Foreach) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Foreach) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	if n.Key != nil {
		n.Key.Walk(v)
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
	}

	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}

	v.LeaveNode(n)
}
