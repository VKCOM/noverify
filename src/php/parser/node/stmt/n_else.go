package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Else node
type Else struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmt         node.Node
	AltSyntax    bool // Whether alternative colon-style syntax is used
}

// NewElse node constructor
func NewElse(Stmt node.Node) *Else {
	return &Else{
		FreeFloating: nil,
		Stmt:         Stmt,
	}
}

// SetPosition sets node position
func (n *Else) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Else) GetPosition() *position.Position {
	return n.Position
}

func (n *Else) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Else) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}

	v.LeaveNode(n)
}
