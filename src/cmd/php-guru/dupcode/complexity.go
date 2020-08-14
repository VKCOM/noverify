package dupcode

import (
	"github.com/VKCOM/noverify/src/ir"
)

// assertComplexity reports whether the given node list reaches the min complexity value.
//
// We don't calculate the precise complexity to avoid full IR traversal for big functions.
// When we reached the min complexity goal, the traversal terminates.
func assertComplexity(list []ir.Node, min uint) bool {
	// Most nodes have the "cost" of at least 1, so
	// if we have more than min nodes, we'll probably will
	// reach the complexity requirement.
	if uint(len(list)) >= min {
		return true
	}

	c := complexityWalker{target: min}
	for _, n := range list {
		n.Walk(&c)
		if c.goalReached() {
			return true
		}
	}
	return c.goalReached()
}

type complexityWalker struct {
	current uint
	target  uint
}

func (c *complexityWalker) goalReached() bool {
	return c.current >= c.target
}

func (c *complexityWalker) LeaveNode(n ir.Node) {}

func (c *complexityWalker) EnterNode(n ir.Node) bool {
	if c.goalReached() {
		return false
	}

	delta := uint(1)

	switch n := n.(type) {
	case *ir.Argument, *ir.ArrayItemExpr:
		delta = 0
	case *ir.IfStmt, *ir.TryStmt:
		delta = 2
	case *ir.ForStmt, *ir.ForeachStmt, *ir.WhileStmt, *ir.DoStmt, *ir.SwitchStmt:
		delta = 3
	case *ir.String:
		delta = uint(len(n.Value)/25) + 1
	case *ir.EncapsedStringPart:
		delta = uint(len(n.Value)/25) + 1
	}

	c.current += delta
	return !c.goalReached()
}
