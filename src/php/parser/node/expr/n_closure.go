package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Closure node
type Closure struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	ReturnsRef    bool
	Static        bool
	PhpDocComment string
	Params        []node.Node
	ClosureUse    *ClosureUse
	ReturnType    node.Node
	Stmts         []node.Node
}

// NewClosure node constructor
func NewClosure(Params []node.Node, ClosureUse *ClosureUse, ReturnType node.Node, Stmts []node.Node, Static bool, ReturnsRef bool, PhpDocComment string) *Closure {
	return &Closure{
		FreeFloating:  nil,
		ReturnsRef:    ReturnsRef,
		Static:        Static,
		PhpDocComment: PhpDocComment,
		Params:        Params,
		ClosureUse:    ClosureUse,
		ReturnType:    ReturnType,
		Stmts:         Stmts,
	}
}

// SetPosition sets node position
func (n *Closure) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Closure) GetPosition() *position.Position {
	return n.Position
}

func (n *Closure) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Closure) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.Params != nil {
		for _, nn := range n.Params {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	if n.ClosureUse != nil {
		n.ClosureUse.Walk(v)
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
