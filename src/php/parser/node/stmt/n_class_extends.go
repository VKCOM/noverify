package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ClassExtends node
type ClassExtends struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	ClassName    node.Node
}

// NewClassExtends node constructor
func NewClassExtends(className node.Node) *ClassExtends {
	return &ClassExtends{
		FreeFloating: nil,
		ClassName:    className,
	}
}

// SetPosition sets node position
func (n *ClassExtends) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ClassExtends) GetPosition() *position.Position {
	return n.Position
}

func (n *ClassExtends) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ClassExtends) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.ClassName != nil {
		n.ClassName.Walk(v)
	}

	v.LeaveNode(n)
}
