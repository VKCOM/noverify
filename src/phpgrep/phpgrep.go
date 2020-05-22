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

type CapturedNode struct {
	Name string
	Node node.Node
}

type MatchData struct {
	Node    node.Node
	Capture []CapturedNode
}

func (data MatchData) CapturedByName(name string) (node.Node, bool) {
	return findNamed(data.Capture, name)
}

// Clone returns a deep copy of m.
func (m *Matcher) Clone() *Matcher {
	return &Matcher{m: m.m}
}

// Match attempts to match n without recursing into it.
//
// Returned match data should only be examined if the
// second return value is true.
func (m *Matcher) Match(n node.Node) (MatchData, bool) {
	var state matcherState
	return m.m.match(&state, n)
}
