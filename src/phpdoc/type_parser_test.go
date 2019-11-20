package phpdoc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TODO: parsing error tests.

func TestTypeParser(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{` int `, `int`},
		{`int`, `int`},
		{`(string)`, `string`},
		{` ( (string))`, `string`},
		{` ((string ) ) `, `string`},
		{`$this`, `$this`},
		{`\A`, `\A`},
		{`\A\B`, `\A\B`},
		{`Foo\Bar`, `Foo\Bar`},

		{`!int`, `!int`},
		{`!(string)`, `!string`},
		{`!((string))`, `!string`},
		{`!!int`, `!!int`},

		{`?int`, `?int`},
		{`?(string)`, `?string`},
		{`?((string))`, `?string`},
		{`??int`, `??int`},

		{`int[]`, `int[]`},
		{`int[][]`, `int[][]`},
		{`(int)[]`, `int[]`},
		{` int [ ] `, `int[]`},
		{` x[] [ ][  ] `, `x[][][]`},

		{`(int|float)[]`, `(int|float)[]`},
		{`(float|int)[]`, `(float|int)[]`},
		{`\A\B|C|int`, `(\A\B|(C|int))`},
		{`int|float`, `(int|float)`},
		{`float|int`, `(float|int)`},
		{`a|b|c`, `(a|(b|c))`},
		{`!a|!b|c`, `(!a|(!b|c))`},
		{`!a|b|!c|d`, `(!a|(b|(!c|d)))`},
		{`!(a|b)|!c|d`, `(!(a|b)|(!c|d))`},
		{`(a|b)|(d|c)`, `((a|b)|(d|c))`},
		{`(a|b)|(d|c|e)|x`, `((a|b)|((d|(c|e))|x))`},
		{`int[]|string[]`, `(int[]|string[])`},
		{`(int[])|string[]`, `(int[]|string[])`},
		{`int[]|(string[])`, `(int[]|string[])`},

		{`a|b&c`, `(a|(b&c))`},
		{`(a|b)&c`, `((a|b)&c)`},
		{`a|(b&c)`, `(a|(b&c))`},
		{`a&b|c`, `(a&(b|c))`},
		{`(a&b)|c`, `((a&b)|c)`},
		{`a&(b|c)`, `(a&(b|c))`},
	}

	var p TypeParser
	for _, test := range tests {
		typ, err := p.ParseType(test.input)
		if err != nil {
			t.Errorf("unexpected error for parse(%q): %v", test.input, err)
			continue
		}
		have := typ.String()
		if have != test.want {
			t.Errorf("result mismatch for parse(%q):\nhave: %s\nwant: %s", test.input, have, test.want)
			t.Logf("%#v", typ)
			continue
		}

		typ2, err := p.ParseType(have)
		if err != nil {
			t.Errorf("re-parse %q error: %v", have, err)
			continue
		}
		have2 := typ2.String()
		if have2 != have {
			t.Errorf("re-parse result mismatch:\nhave: %s\nwant: %s", have2, have)
			continue
		}
		if diff := cmp.Diff(typ, typ2); diff != "" {
			t.Errorf("re-parse result mismatch: %s", diff)
		}
	}
}
