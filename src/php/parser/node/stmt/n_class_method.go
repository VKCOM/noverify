package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ClassMethod node
type ClassMethod struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	PhpDocComment string
	MethodName    *node.Identifier
	Modifiers     []*node.Identifier
	Params        []node.Node
	ReturnType    node.Node
	Stmt          node.Node
}

// NewClassMethod node constructor
func NewClassMethod(MethodName *node.Identifier, Modifiers []*node.Identifier, ReturnsRef bool, Params []node.Node, ReturnType node.Node, Stmt node.Node, PhpDocComment string) *ClassMethod {
	return &ClassMethod{
		FreeFloating:  nil,
		ReturnsRef:    ReturnsRef,
		PhpDocComment: PhpDocComment,
		MethodName:    MethodName,
		Modifiers:     Modifiers,
		Params:        Params,
		ReturnType:    ReturnType,
		Stmt:          Stmt,
	}
}

// SetPosition sets node position
func (n *ClassMethod) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ClassMethod) GetPosition() *position.Position {
	return n.Position
}

func (n *ClassMethod) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ClassMethod) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.MethodName != nil {
		n.MethodName.Walk(v)
	}

	if n.Modifiers != nil {
		for _, nn := range n.Modifiers {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.Params != nil {
		for _, nn := range n.Params {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.ReturnType != nil {
		n.ReturnType.Walk(v)
	}

	if n.Stmt != nil {
		n.Stmt.Walk(v)
	}

	v.LeaveNode(n)
}
