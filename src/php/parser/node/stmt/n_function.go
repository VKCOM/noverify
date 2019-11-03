package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Function node
type Function struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	PhpDocComment string
	FunctionName  *node.Identifier
	Params        []node.Node
	ReturnType    node.Node
	Stmts         []node.Node
}

// NewFunction node constructor
func NewFunction(FunctionName *node.Identifier, ReturnsRef bool, Params []node.Node, ReturnType node.Node, Stmts []node.Node, PhpDocComment string) *Function {
	return &Function{
		FreeFloating:  nil,
		ReturnsRef:    ReturnsRef,
		PhpDocComment: PhpDocComment,
		FunctionName:  FunctionName,
		Params:        Params,
		ReturnType:    ReturnType,
		Stmts:         Stmts,
	}
}

// SetPosition sets node position
func (n *Function) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Function) GetPosition() *position.Position {
	return n.Position
}

func (n *Function) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Function) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.FunctionName != nil {
		n.FunctionName.Walk(v)
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

	if n.Stmts != nil {
		for _, nn := range n.Stmts {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
