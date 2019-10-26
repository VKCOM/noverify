// Package walker declares walking behavior
package walker

// Walkable interface
//
// Every node must implement this interface
type Walkable interface {
	Walk(v Visitor)
}

// Visitor interface
type Visitor interface {
	EnterNode(w Walkable) bool
	LeaveNode(w Walkable)
}
