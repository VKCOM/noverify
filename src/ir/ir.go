// Package ir declares intermediate representation type suitable for the analysis.
//
// IR is generated from the AST, see ir/irconv package.
package ir

import (
	"github.com/VKCOM/noverify/src/php/parser/freefloating"
)

//go:generate go run ./codegen

// Node is a type that is implemented by all IR types.
// node_types.go contains all implementations.
type Node interface {
	Walk(Visitor)
	GetFreeFloating() *freefloating.Collection
	IterateTokens(func(*Token) bool)
}

// Visitor is an interface for basic IR nodes traversal.
type Visitor interface {
	EnterNode(Node) bool
	LeaveNode(Node)
}
