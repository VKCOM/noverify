package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// AltElse node
type AltElse struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Stmt         node.Node
}

// NewAltElse node constructor
func NewAltElse(Stmt node.Node) *AltElse {
	return &AltElse{
		FreeFloating: nil,
		Stmt:         Stmt,
	}
}

// SetPosition sets node position
func (n *AltElse) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *AltElse) GetPosition() *position.Position {
	return n.Position
}

func (n *AltElse) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Attributes returns node attributes as map
func (n *AltElse) Attributes() map[string]interface{} {
	return nil
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *AltElse) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Stmt != nil {
		v.EnterChildNode("Stmt", n)
		n.Stmt.Walk(v)
		v.LeaveChildNode("Stmt", n)
	}

	v.LeaveNode(n)
}
