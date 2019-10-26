package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// HaltCompiler node
type HaltCompiler struct {
	FreeFloating freefloating.Collection
	Position     *position.Position
}

// NewHaltCompiler node constructor
func NewHaltCompiler() *HaltCompiler {
	return &HaltCompiler{}
}

// SetPosition sets node position
func (n *HaltCompiler) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *HaltCompiler) GetPosition() *position.Position {
	return n.Position
}

func (n *HaltCompiler) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *HaltCompiler) Walk(v walker.Visitor) {
	if v.EnterNode(n) == false {
		return
	}

	v.LeaveNode(n)
}
