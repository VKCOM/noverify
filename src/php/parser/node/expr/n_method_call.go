package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// MethodCall node
type MethodCall struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Variable     node.Node
	Method       node.Node
	ArgumentList *node.ArgumentList
}

// NewMethodCall node constructor
func NewMethodCall(Variable node.Node, Method node.Node, ArgumentList *node.ArgumentList) *MethodCall {
	return &MethodCall{
		FreeFloating: nil,
		Variable:     Variable,
		Method:       Method,
		ArgumentList: ArgumentList,
	}
}

// SetPosition sets node position
func (n *MethodCall) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *MethodCall) GetPosition() *position.Position {
	return n.Position
}

func (n *MethodCall) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *MethodCall) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Variable != nil {
		n.Variable.Walk(v)
	}

	if n.Method != nil {
		n.Method.Walk(v)
	}

	if n.ArgumentList != nil {
		n.ArgumentList.Walk(v)
	}

	v.LeaveNode(n)
}
