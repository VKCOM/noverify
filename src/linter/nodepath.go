package linter

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
)

type NodePath struct {
	stack []ir.Node
}

func newNodePath() NodePath {
	return NodePath{stack: make([]ir.Node, 0, 20)}
}

func (p NodePath) String() string {
	parts := make([]string, len(p.stack))
	for i, n := range p.stack {
		parts[i] = fmt.Sprintf("%T", n)
	}
	return strings.Join(parts, "/")
}

func (p NodePath) Parent() ir.Node {
	return p.NthParent(1)
}

func (p NodePath) Current() ir.Node {
	return p.NthParent(0)
}

func (p NodePath) Conditional() bool {
	return p.ConditionalUntil(p.Current())
}

func (p NodePath) NthParent(n int) ir.Node {
	index := len(p.stack) - n - 1
	if index >= 0 {
		return p.stack[index]
	}
	return nil
}

func (p NodePath) ConditionalUntil(end ir.Node) bool {
	// TODO: can we report if statement as unconditioncal?
	// stmt.If is pushed before we handle condition,
	// so it will occure as a path element.

	for _, n := range p.stack {
		if n == end {
			break
		}

		switch n.(type) {
		// Obvious branching.
		case *ir.IfStmt, *ir.ElseIfStmt, *ir.TernaryExpr:
			return true
		// Else executes only if condition failed.
		case *ir.ElseStmt:
			return true
		// Loops that can be executed 0 times.
		case *ir.ForStmt, *ir.ForeachStmt, *ir.WhileStmt:
			return true
		case *ir.CatchStmt:
			return true
		}
	}

	return false
}

func (p *NodePath) push(n ir.Node) {
	p.stack = append(p.stack, n)
}

func (p *NodePath) pop() {
	p.stack = p.stack[:len(p.stack)-1]
}
