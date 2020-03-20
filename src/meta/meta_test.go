package meta

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/php/parser/node"
	"github.com/VKCOM/noverify/src/php/parser/node/name"
)

func TestNameEquals(t *testing.T) {
	tests := []struct {
		x    string
		y    string
		want bool
	}{
		{`foo`, `foo`, true},
		{`foo\bar2`, `foo\bar2`, true},
		{`foo\bar2\BazBaz`, `foo\bar2\BazBaz`, true},
		{`foo \ bar`, `foo\bar`, true},
		{`a\b\c\d`, `a\b\c\d`, true},

		{`a\b`, `a\bb`, false},
		{`a b`, `a\b`, false},
		{`a b\`, `a\b`, false},
		{`first`, ``, false},
		{`first\second`, `first`, false},
		{`first`, `first\second`, false},
		{`first\first`, `first\second`, false},
		{`first\second`, `firstsecond`, false},
		{`firstsecond`, `first\second`, false},
		{`a\b\c`, `a\b\x`, false},
	}

	makeNameNode := func(s string) *name.Name {
		parts := strings.Split(s, `\`)
		nm := &name.Name{Parts: make([]node.Node, len(parts))}
		for i := range parts {
			nm.Parts[i] = &name.NamePart{Value: strings.TrimSpace(parts[i])}
		}
		return nm
	}

	for _, test := range tests {
		x := makeNameNode(test.x)
		y := test.y
		have := NameEquals(x, y)
		if have != test.want {
			t.Errorf("NameEquals(%q, %q): have %v, want %v",
				test.x, test.y, have, test.want)
		}
	}
}

func BenchmarkNameEquals(b *testing.B) {
	var theName1 = &name.Name{
		Parts: []node.Node{
			&name.NamePart{Value: `method_exists`},
		},
	}
	var theName3 = &name.Name{
		Parts: []node.Node{
			&name.NamePart{Value: `a`},
			&name.NamePart{Value: `b`},
			&name.NamePart{Value: `c`},
		},
	}

	b.Run("1part", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NameEquals(theName1, `method_exists`)
		}
	})
	b.Run("3parts", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NameEquals(theName3, `a\b\c`)
		}
	})
}
