package dupcode

import (
	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irutil"
)

func findNode(root ir.Node, pred func(n ir.Node) bool) bool {
	found := false
	irutil.Inspect(root, func(n ir.Node) bool {
		if found {
			return false
		}
		if pred(n) {
			found = true
		}
		if v, ok := n.(*ir.SimpleVar); ok {
			if v.Name == "this" {
				found = true
			}
		}
		return !found
	})
	if found {
		return true
	}
	return found
}

func hasModifier(list []*ir.Identifier, key string) bool {
	for _, x := range list {
		if x.Value == key {
			return true
		}
	}
	return false
}
