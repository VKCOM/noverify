package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Try node
type Try struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []node.Node
	Catches      []node.Node
	Finally      node.Node
}

// NewTry node constructor
func NewTry(Stmts []node.Node, Catches []node.Node, Finally node.Node) *Try {
	return &Try{
		FreeFloating: nil,
		Stmts:        Stmts,
		Catches:      Catches,
		Finally:      Finally,
	}
}

// SetPosition sets node position
func (n *Try) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Try) GetPosition() *position.Position {
	return n.Position
}

func (n *Try) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Try) Walk(v walker.Visitor) {
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

	if n.Catches != nil {
		for _, nn := range n.Catches {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.Finally != nil {
		n.Finally.Walk(v)
	}

	v.LeaveNode(n)
}
