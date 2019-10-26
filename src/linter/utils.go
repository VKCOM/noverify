package linter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/printer"
	"github.com/VKCOM/noverify/src/php/parser/walker"
	"github.com/VKCOM/noverify/src/solver"
)

// FmtNode is used for debug purposes and returns string representation of a specified node.
func FmtNode(n node.Node) string {
	var b bytes.Buffer
	printer.NewPrettyPrinter(&b, " ").Print(n)
	return b.String()
}

// FlagsToString is designed for debugging flags.
func FlagsToString(f int) string {
	var res []string

	if (f & FlagReturn) == FlagReturn {
		res = append(res, "Return")
	}

	if (f & FlagDie) == FlagDie {
		res = append(res, "Die")
	}

	if (f & FlagThrow) == FlagThrow {
		res = append(res, "Throw")
	}

	if (f & FlagBreak) == FlagBreak {
		res = append(res, "Break")
	}

	return "Exit flags: [" + strings.Join(res, ", ") + "], digits: " + fmt.Sprintf("%d", f)
}

func haveMagicMethod(class string, methodName string) bool {
	_, _, ok := solver.FindMethod(class, methodName)
	return ok
}

func isQuote(r rune) bool {
	return r == '"' || r == '\''
}

// walkNode is a convenience wrapper for EnterNode-only traversals.
// It gives a way to traverse a node without defining a new kind of walker.
//
// enterNode function is called in place where EnterNode method would be called.
// If n is nil, no traversal is performed.
func walkNode(n node.Node, enterNode func(walker.Walkable) bool) {
	if n == nil {
		return
	}
	v := nodeVisitor{enterNode: enterNode}
	n.Walk(v)
}

type nodeVisitor struct {
	enterNode func(walker.Walkable) bool
}

func (v nodeVisitor) EnterChildNode(key string, w walker.Walkable) {}
func (v nodeVisitor) LeaveChildNode(key string, w walker.Walkable) {}
func (v nodeVisitor) EnterChildList(key string, w walker.Walkable) {}
func (v nodeVisitor) LeaveChildList(key string, w walker.Walkable) {}
func (v nodeVisitor) LeaveNode(w walker.Walkable)                  {}

func (v nodeVisitor) EnterNode(w walker.Walkable) bool {
	return v.enterNode(w)
}
