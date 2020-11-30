package phpgrep

import (
	"github.com/VKCOM/noverify/src/ir"
)

// Compiler creates matcher objects out of the string patterns.
type Compiler struct {
	// CaseSensitive option specifies whether compiled patterns
	// should match identifiers in a strictly case-sensitive manner.
	//
	// In PHP, f() and F() refer to the same function `f`, but if
	// case sensitivity is set to true, compiled matcher will reject
	// any spelling mismatches.
	CaseSensitive bool
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
	Node ir.Node
}

type MatchData struct {
	Node    ir.Node
	Capture []CapturedNode
}

func (data MatchData) CapturedByName(name string) (ir.Node, bool) {
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
func (m *Matcher) Match(n ir.Node) (MatchData, bool) {
	var state matcherState
	return m.m.match(&state, n)
}
