// Package ir declares intermediate representation type suitable for the analysis.
//
// IR is generated from the AST, see ir/irconv package.
package ir

import (
	"github.com/VKCOM/php-parser/pkg/token"
)

//go:generate go run ./codegen

// Node is a type that is implemented by all IR types.
// node_types.go contains all implementations.
type Node interface {
	Parent() Node
	Walk(Visitor)
	IterateTokens(func(*token.Token) bool)
}

// Visitor is an interface for basic IR nodes traversal.
type Visitor interface {
	EnterNode(Node) bool
	LeaveNode(Node)
}
