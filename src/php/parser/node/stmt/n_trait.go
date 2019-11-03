package stmt

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Trait node
type Trait struct {
	FreeFloating  freefloating.Collection
	Position      *position.Position
	PhpDocComment string
	TraitName     *node.Identifier
	Stmts         []node.Node
}

// NewTrait node constructor
func NewTrait(TraitName *node.Identifier, Stmts []node.Node, PhpDocComment string) *Trait {
	return &Trait{
		FreeFloating:  nil,
		PhpDocComment: PhpDocComment,
		TraitName:     TraitName,
		Stmts:         Stmts,
	}
}

// SetPosition sets node position
func (n *Trait) SetPosition(p *position.Position) {
	n.Position = p
}

// GetPosition returns node positions
func (n *Trait) GetPosition() *position.Position {
	return n.Position
}

func (n *Trait) GetFreeFloating() *freefloating.Collection {
	return &n.FreeFloating
}

// Walk traverses nodes
// Walk is invoked recursively until v.EnterNode returns true
func (n *Trait) Walk(v walker.Visitor) {
	if !v.EnterNode(n) {
		return
	}

	if n.TraitName != nil {
		n.TraitName.Walk(v)
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
