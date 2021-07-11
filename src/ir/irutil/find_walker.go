package irutil

import (
	"github.com/VKCOM/noverify/src/ir"
)

type findPredicate func(what ir.Node, cur ir.Node) bool

// findWalker structure implements a walker to find
// a specific node in a subtree.
type findWalker struct {
	what  ir.Node
	where ir.Node
	found bool

	withPredicate bool
	predicate     findPredicate
}

// newFindWalker returns a walker with a predicate.
//
// If the predicate returns true, the search stops.
func newFindWalkerWithPredicate(what ir.Node, where ir.Node, pred findPredicate) *findWalker {
	return &findWalker{
		what:          what,
		where:         where,
		withPredicate: true,
		predicate:     pred,
	}
}

func (w *findWalker) EnterNode(n ir.Node) (res bool) {
	if w.withPredicate {
		if w.predicate(w.what, n) {
			w.found = true
			return false
		}
	}

	if NodeEqual(n, w.what) {
		w.found = true
		return false
	}
	return true
}

func (w *findWalker) LeaveNode(ir.Node) {}
