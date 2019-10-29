package phpgrep

import (
	"github.com/VKCOM/noverify/src/php/parser/node"
)

// Compiler creates matcher objects out of the string patterns.
type Compiler struct {
}

// Compile compiler a given pattern into a matcher.
func (c *Compiler) Compile(pattern []byte) (*Matcher, error) {
	return compile(c, pattern)
}

// Matcher is a compiled pattern that can be used for PHP code search.
type Matcher struct {
	m matcher
}

type MatchData struct {
	Node  node.Node
	Named map[string]node.Node
}

// Clone returns a deep copy of m.
func (m *Matcher) Clone() *Matcher {
	return &Matcher{m: m.m}
}

// Match reports whether given PHP code matches the bound pattern.
//
// For malformed inputs (like code with syntax errors), returns false.
func (m *Matcher) Match(root node.Node) bool {
	return m.m.matchAST(root)
}

func (m *Matcher) Find(root node.Node, callback func(*MatchData) bool) {
	m.m.findAST(root, callback)
}
