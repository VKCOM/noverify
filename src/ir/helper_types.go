package ir

// Helper types are not real nodes, they're usually used
// to express some structure that is common between several nodes.
//
// In other words, structs defined in this file do not implement the Node interface.

// Class is a common shape between the ClassStmt and AnonClassExpr.
// It doesn't include positions/freefloating info.
type Class struct {
	PhpDocComment string
	Extends       *ClassExtendsStmt
	Implements    *ClassImplementsStmt
	Stmts         []Node
}
