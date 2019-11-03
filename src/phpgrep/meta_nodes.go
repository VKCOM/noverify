package phpgrep

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
	"github.com/VKCOM/noverify/src/php/parser/position"
	"github.com/VKCOM/noverify/src/php/parser/walker"
)

type metaNode struct {
	name string
}

func (metaNode) Walk(v walker.Visitor)                     {}
func (metaNode) SetPosition(p *position.Position)          {}
func (metaNode) GetPosition() *position.Position           { return nil }
func (metaNode) GetFreeFloating() *freefloating.Collection { return nil }

type (
	anyConst struct{ metaNode }
	anyVar   struct{ metaNode }
	anyInt   struct{ metaNode }
	anyFloat struct{ metaNode }
	anyStr   struct{ metaNode }
	anyNum   struct{ metaNode }
	anyExpr  struct{ metaNode }
	anyFunc  struct{ metaNode }
)
