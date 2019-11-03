package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// StmtList node
type StmtList struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmts        []node.Node
}

// NewStmtList node constructor
func NewStmtList(Stmts []node.Node) *StmtList {
	return &StmtList{
		FreeFloating: nil,
		Stmts:        Stmts,
	}
}

// SetPosition sets node position
func (n *StmtList) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *StmtList) GetPosition() *position.Position {
	return n.Position
}

func (n *StmtList) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *StmtList) Walk(v walker.Visitor) {
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

	v.LeaveNode(n)
}
