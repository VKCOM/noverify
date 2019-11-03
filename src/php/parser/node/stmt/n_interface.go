package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Interface node
type Interface struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	InterfaceName *node.Identifier
	Extends       *InterfaceExtends
	Stmts         []node.Node
}

// NewInterface node constructor
func NewInterface(InterfaceName *node.Identifier, Extends *InterfaceExtends, Stmts []node.Node, PhpDocComment string) *Interface {
	return &Interface{
		FreeFloating:  nil,
		PhpDocComment: PhpDocComment,
		InterfaceName: InterfaceName,
		Extends:       Extends,
		Stmts:         Stmts,
	}
}

// SetPosition sets node position
func (n *Interface) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Interface) GetPosition() *position.Position {
	return n.Position
}

func (n *Interface) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Interface) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.InterfaceName != nil {
		n.InterfaceName.Walk(v)
	}

	if n.Extends != nil {
		n.Extends.Walk(v)
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
