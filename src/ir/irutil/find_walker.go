package irutil

import (
	"github.com/VKCOM/noverify/src/ir"
)

type findWalker struct {
	what  ir.Node
	where ir.Node
	Found bool

	withPredicate bool
	predicate     func(what ir.Node, cur ir.Node) bool
}

func newFindWalker(what ir.Node, where ir.Node) *findWalker {
	return &findWalker{
		what:  what,
		where: where,
	}
}

func newFindWalkerWithPredicate(what ir.Node, where ir.Node, pred func(what ir.Node, cur ir.Node) bool) *findWalker {
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
			w.Found = true
			return false
		}
	}

	if NodeEqual(n, w.what) {
		w.Found = true
		return false
	}
	return true
}

func (w *findWalker) LeaveNode(n ir.Node) {}
