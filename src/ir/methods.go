package ir

import (
	"strings"
)

// IsFullyQualified reports whether the name is fully qualified.
// FQN don't need any further resolution.
func (n *Name) IsFullyQualified() bool {
	return strings.HasPrefix(n.Value, `\`)
}

// NumParts reports number of the name parts.
func (n *Name) NumParts() int {
	s := n.Value
	if n.IsFullyQualified() {
		s = s[len(`\`):]
	}
	return strings.Count(s, `\`) + 1
}

// HeadTail is an efficient way to get <FirstPart, RestParts> pair.
func (n *Name) HeadTail() (head, tail string) {
	s := n.Value
	if n.IsFullyQualified() {
		s = s[len(`\`):]
	}
	slash := strings.IndexByte(s, '\\')
	if slash == -1 {
		return s, ""
	}
	return s[:slash], s[len(`\`)+slash:]
}

// LastPart returns only the last name part.
// If name contains only one part, that part is returned.
func (n *Name) LastPart() string {
	s := n.Value
	if n.IsFullyQualified() {
		s = s[len(`\`):]
	}
	lastSlash := strings.LastIndexByte(s, '\\')
	if lastSlash == -1 {
		return s
	}
	return s[len(`\`)+lastSlash:]
}

// FirstPart returns only the first name part.
// If name contains only one part, that part is returned.
func (n *Name) FirstPart() string {
	s := n.Value
	if n.IsFullyQualified() {
		s = s[len(`\`):]
	}
	slash := strings.IndexByte(s, '\\')
	if slash == -1 {
		return s
	}
	return s[:slash]
}

// RestParts returns all but first name parts.
// If name contains only one part, empty string is returned.
func (n *Name) RestParts() string {
	s := n.Value
	if n.IsFullyQualified() {
		s = s[len(`\`):]
	}
	slash := strings.IndexByte(s, '\\')
	if slash == -1 {
		return ""
	}
	return s[len(`\`)+slash:]
}

// Arg returns the ith argument.
func (n *FunctionCallExpr) Arg(i int) *Argument { return n.Args[i].(*Argument) }

// Arg returns the ith argument.
func (n *MethodCallExpr) Arg(i int) *Argument { return n.Args[i].(*Argument) }

// Arg returns the ith argument.
func (n *NewExpr) Arg(i int) *Argument { return n.Args[i].(*Argument) }

// Arg returns the ith argument.
func (n *StaticCallExpr) Arg(i int) *Argument { return n.Args[i].(*Argument) }

// Arg returns the ith argument.
func (n *ClassStmt) Arg(i int) *Argument { return n.Args[i].(*Argument) }
