package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// If node
type If struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Cond         node.Node
	Stmt         node.Node
	ElseIf       []node.Node // Always []*ElseIf
	Else         node.Node
	AltSyntax    bool // Whether alternative colon-style syntax is used
}

// NewIf node constructor
func NewIf(Cond node.Node, Stmt node.Node, ElseIf []node.Node, Else node.Node) *If {
	return &If{
		FreeFloating: nil,
		Cond:         Cond,
		Stmt:         Stmt,
		ElseIf:       ElseIf,
		Else:         Else,
	}
}

// SetPosition sets node position
func (n *If) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *If) GetPosition() *position.Position {
	return n.Position
}

func (n *If) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// AddElseIf add ElseIf node and returns AltIf node
func (n *If) AddElseIf(ElseIf node.Node) node.Node {
	n.ElseIf = append(n.ElseIf, ElseIf)
	return n
}

// SetElse set Else node and returns AltIf node
func (n *If) SetElse(Else node.Node) node.Node {
	n.Else = Else

	return n
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *If) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Cond != nil {
		n.Cond.Walk(v)
	}

	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}

	if n.ElseIf != nil {
		for _, nn := range n.ElseIf {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.Else != nil {
		n.Else.Walk(v)
	}

	v.LeaveNode(n)
}
