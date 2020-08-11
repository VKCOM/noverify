package irgen

import (
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
)

func fullyQualifiedToString(n *name.FullyQualified) string {
	s := make([]string, 1, len(n.Parts)+1)
	for _, v := range n.Parts {
		s = append(s, v.(*name.NamePart).Value)
	}
	return strings.Join(s, `\`)
}

// namePartsToString converts slice of *name.NamePart to string
func namePartsToString(parts []node.Node) string {
	s := make([]string, 0, len(parts))
	for _, v := range parts {
		s = append(s, v.(*name.NamePart).Value)
	}
	return strings.Join(s, `\`)
}
