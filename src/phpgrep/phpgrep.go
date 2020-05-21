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

// Find executed a callback for every match inside root.
// If callback returns false, currently matched node children are not visited.
func (m *Matcher) Find(root node.Node, callback func(*MatchData) bool) {
	m.m.findAST(root, callback)
}

// Match attempts to match n without recursing into it.
//
// Returned match data should only be examined if the
// second return value is true.
func (m *Matcher) Match(n node.Node) (MatchData, bool) {
	found := m.m.match(n)
	return m.m.data, found
}
