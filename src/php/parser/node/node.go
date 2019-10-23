package node

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

// Node interface
type Node interface {
	walker.Walkable
	Attributes() map[string]interface{} // Attributes returns node attributes as map
	SetPosition(p *position.Position)
	GetPosition() *position.Position
	GetFreeFloating() *freefloating.Collection
}
