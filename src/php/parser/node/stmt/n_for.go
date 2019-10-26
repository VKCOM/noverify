package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// For node
type For struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Init         []node.Node
	Cond         []node.Node
	Loop         []node.Node
	Stmt         node.Node
	AltSyntax    bool // Whether alternative colon-style syntax is used
}

// NewFor node constructor
func NewFor(Init []node.Node, Cond []node.Node, Loop []node.Node, Stmt node.Node) *For {
	return &For{
		FreeFloating: nil,
		Init:         Init,
		Cond:         Cond,
		Loop:         Loop,
		Stmt:         Stmt,
	}
}

// SetPosition sets node position
func (n *For) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *For) GetPosition() *position.Position {
	return n.Position
}

func (n *For) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *For) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Init != nil {
		for _, nn := range n.Init {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.Cond != nil {
		for _, nn := range n.Cond {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.Loop != nil {
		for _, nn := range n.Loop {
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
