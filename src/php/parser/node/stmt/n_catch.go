package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Catch node
type Catch struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Types        []node.Node
	Variable     *node.SimpleVar
	Stmts        []node.Node
}

// NewCatch node constructor
func NewCatch(Types []node.Node, Variable *node.SimpleVar, Stmts []node.Node) *Catch {
	return &Catch{
		FreeFloating: nil,
		Types:        Types,
		Variable:     Variable,
		Stmts:        Stmts,
	}
}

// SetPosition sets node position
func (n *Catch) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Catch) GetPosition() *position.Position {
	return n.Position
}

func (n *Catch) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Catch) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Types != nil {
		for _, nn := range n.Types {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
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
