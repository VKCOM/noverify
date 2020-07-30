package linter

import (
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/expr"
	"github.com/VKCOM/noverify/src/php/parser/node/stmt"
)

type NodePath struct {
	stack []node.Node
}

func newNodePath() NodePath {
	return NodePath{stack: make([]node.Node, 0, 20)}
}

func (p NodePath) String() string {
	parts := make([]string, len(p.stack))
	for i, n := range p.stack {
		parts[i] = fmt.Sprintf("%T", n)
	}
	return strings.Join(parts, "/")
}

func (p NodePath) Parent() node.Node {
	return p.NthParent(1)
}

func (p NodePath) Current() node.Node {
	return p.NthParent(0)
}

func (p NodePath) Conditional() bool {
	return p.ConditionalUntil(p.Current())
}

func (p NodePath) NthParent(n int) node.Node {
	index := len(p.stack) - n - 1
	if index >= 0 {
		return p.stack[index]
	}
	return nil
}

func (p NodePath) ConditionalUntil(end node.Node) bool {
	// TODO: can we report if statement as unconditioncal?
	// stmt.If is pushed before we handle condition,
	// so it will occure as a path element.

	for _, n := range p.stack {
		if n == end {
			break
		}

		switch n.(type) {
		// Obvious branching.
		case *stmt.If, *stmt.ElseIf, *expr.Ternary:
			return true
		// Else executes only if condition failed.
		case *stmt.Else:
			return true
		// Loops that can be executed 0 times.
		case *stmt.For, *stmt.Foreach, *stmt.While:
			return true
		case *stmt.Catch:
			return true
		}
	}

	return false
}

func (p *NodePath) push(n node.Node) {
	p.stack = append(p.stack, n)
}

func (p *NodePath) pop() {
	p.stack = p.stack[:len(p.stack)-1]
}
