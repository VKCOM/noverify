package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// FunctionCall node
type FunctionCall struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Function     node.Node
	ArgumentList *node.ArgumentList
}

// NewFunctionCall node constructor
func NewFunctionCall(Function node.Node, ArgumentList *node.ArgumentList) *FunctionCall {
	return &FunctionCall{
		FreeFloating: nil,
		Function:     Function,
		ArgumentList: ArgumentList,
	}
}

// SetPosition sets node position
func (n *FunctionCall) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *FunctionCall) GetPosition() *position.Position {
	return n.Position
}

func (n *FunctionCall) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *FunctionCall) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Function != nil {
		n.Function.Walk(v)
	}

	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}

	v.LeaveNode(n)
}
