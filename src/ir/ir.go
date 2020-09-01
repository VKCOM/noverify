// Package ir declares intermediate representation type suitable for the analysis.
//
// IR is generated from the AST, see ir/irconv package.
package ir

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
)

type Visitor interface {
	EnterNode(Node) bool
	LeaveNode(Node)
}

type Node interface {
	Walk(Visitor)
	GetFreeFloating() *freefloating.Collection
}
