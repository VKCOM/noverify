package irutil

import (
	"github.com/VKCOM/noverify/src/ir"
)

type findPredicate func(node ir.Node) bool

// findWalker structure implements a walker to find
// a specific node in a subtree.
type findWalker struct {
	where ir.Node
	res   ir.Node

	predicate findPredicate
}

// newFindWalker returns a walker with a predicate.
//
// If the predicate returns true, the search stops.
func newFindWalkerWithPredicate(where ir.Node, pred findPredicate) *findWalker {
	return &findWalker{
		where:     where,
		predicate: pred,
	}
}

func (w *findWalker) EnterNode(n ir.Node) (res bool) {
	if w.predicate(n) {
		w.res = n
		return false
	}

	return true
}

func (w *findWalker) LeaveNode(ir.Node) {}
