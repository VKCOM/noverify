package meta

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/ir"
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

	makeNameNode := func(s string) *ir.Name {
		parts := strings.Split(s, `\`)
		nm := &ir.Name{Parts: make([]ir.Node, len(parts))}
		for i := range parts {
			nm.Parts[i] = &ir.NamePart{Value: strings.TrimSpace(parts[i])}
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
	var theName1 = &ir.Name{
		Parts: []ir.Node{
			&ir.NamePart{Value: `method_exists`},
		},
	}
	var theName3 = &ir.Name{
		Parts: []ir.Node{
			&ir.NamePart{Value: `a`},
			&ir.NamePart{Value: `b`},
			&ir.NamePart{Value: `c`},
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
