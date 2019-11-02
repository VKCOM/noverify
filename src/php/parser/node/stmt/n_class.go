package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Class node
type Class struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	ClassName     *node.Identifier
	Modifiers     []*node.Identifier
	ArgumentList  *node.ArgumentList
	Extends       *ClassExtends
	Implements    *ClassImplements
	Stmts         []node.Node
}

// NewClass node constructor
func NewClass(ClassName *node.Identifier, Modifiers []*node.Identifier, ArgumentList *node.ArgumentList, Extends *ClassExtends, Implements *ClassImplements, Stmts []node.Node, PhpDocComment string) *Class {
	return &Class{
		FreeFloating:  nil,
		PhpDocComment: PhpDocComment,
		ClassName:     ClassName,
		Modifiers:     Modifiers,
		ArgumentList:  ArgumentList,
		Extends:       Extends,
		Implements:    Implements,
		Stmts:         Stmts,
	}
}

// SetPosition sets node position
func (n *Class) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Class) GetPosition() *position.Position {
	return n.Position
}

func (n *Class) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Class) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.ClassName != nil {
		n.ClassName.Walk(v)
	}

	if n.Modifiers != nil {
		for _, nn := range n.Modifiers {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}

	if n.Extends != nil {
		n.Extends.Walk(v)
	}

	if n.Implements != nil {
		n.Implements.Walk(v)
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
