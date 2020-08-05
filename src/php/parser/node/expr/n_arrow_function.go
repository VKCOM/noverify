package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ArrowFunction node
type ArrowFunction struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	Static        bool
	PhpDocComment string
	Params        []node.Node
	ReturnType    node.Node
	Expr          node.Node
}

// NewArrowFunction node constructor
func NewArrowFunction(Params []node.Node, ReturnType node.Node, Stmt node.Node, Static bool, ReturnsRef bool, PhpDocComment string) *ArrowFunction {
	return &ArrowFunction{
		FreeFloating:  nil,
		ReturnsRef:    ReturnsRef,
		Static:        Static,
		PhpDocComment: PhpDocComment,
		Params:        Params,
		ReturnType:    ReturnType,
		Expr:          Stmt,
	}
}

// SetPosition sets node position
func (n *ArrowFunction) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ArrowFunction) GetPosition() *position.Position {
	return n.Position
}

func (n *ArrowFunction) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ArrowFunction) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
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

	if n.Expr != nil {
		n.Expr.Walk(v)
	}

	v.LeaveNode(n)
}
