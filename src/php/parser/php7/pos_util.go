package php7

import (
	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/position"
)

var badPos = &position.Position{
	StartLine: -1,
	StartPos:  -1,
	EndLine:   -1,
	EndPos:    -1,
}

func nodePos(n node.Node) *position.Position {
	if n == nil {
		return badPos
	}

	if pos := n.GetPosition(); pos != nil {
		return pos
	}
	return badPos
}

func identListStartPos(l []*node.Identifier) *position.Position {
	if len(l) == 0 {
		return badPos
	}

	return nodePos(l[0])
}
