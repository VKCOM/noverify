package expr

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// ShellExec node
type ShellExec struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
	Parts        []node.Node
}

// NewShellExec node constructor
func NewShellExec(Parts []node.Node) *ShellExec {
	return &ShellExec{
		FreeFloating: nil,
		Parts:        Parts,
	}
}

// SetPosition sets node position
func (n *ShellExec) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *ShellExec) GetPosition() *position.Position {
	return n.Position
}

func (n *ShellExec) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *ShellExec) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	if n.Parts != nil {
		for _, nn := range n.Parts {
			if nn != nil {
				nn.Walk(v)
			}
		}
	}

	v.LeaveNode(n)
}
