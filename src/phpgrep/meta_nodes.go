package phpgrep

import (
	"github.com/z7zmey/php-parser/pkg/token"

	"github.com/VKCOM/noverify/src/ir"
)

type metaNode struct {
	name string
}

func (metaNode) Walk(v ir.Visitor)                           {}
func (metaNode) IterateTokens(func(token *token.Token) bool) {}

type (
	anyConst struct{ metaNode }
	anyVar   struct{ metaNode }
	anyInt   struct{ metaNode }
	anyFloat struct{ metaNode }
	anyStr   struct{ metaNode }
	anyStr1  struct{ metaNode }
	anyNum   struct{ metaNode }
	anyExpr  struct{ metaNode }
	anyCall  struct{ metaNode }
	anyFunc  struct{ metaNode }
)
