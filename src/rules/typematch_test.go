package rules

import (
	"testing"

	"github.com/VKCOM/noverify/src/phpdoc"
)

func TestTypeFilter(t *testing.T) {
	type testCase struct {
		dst string
		val string
	}

	matchingTests := []testCase{
		{`array`, `mixed[]`},
		{`array`, `int[]`},
		{`array`, `\Foo[]`},

		{`object`, `object`},
		{`object`, `\Foo`},
		{`object`, `\Foo\Bar`},

		{`!int`, `string`},
		{`!int`, `mixed`},
		{`!array`, `int`},
		{`!array`, `string`},

		{`int[]`, `int[]`},

		{`int`, `(int)`},
		{`(int)`, `int`},
		{`(int)`, `((int))`},

		{`int|float`, `int`},
		{`int|float`, `float`},
		{`float|int`, `int`},
		{`int|float`, `float`},

		{`!(int|float)`, `string`},
		{`!(int|float)`, `int[]`},
		{`!(object|array)`, `int`},
		{`!(object|array)`, `string`},

		{`a|b`, `a|b`},
		{`object|array`, `\Foo|int[]`},
		{`object|array`, `object|float[]`},

		{`a|b|c`, `a|b|c`},
		{`a|c|b`, `a|b|c`},
		{`c|b|a`, `a|b|c`},
	}

	nonMatchingTests := []testCase{
		{`array`, `int`},
		{`array`, `mixed`},
		{`array`, `\Foo`},

		{`object`, `int`},
		{`object`, `\Foo[]`},
		{`object`, `mixed`},

		{`!int`, `int`},
		{`!array`, `mixed[]`},
		{`!array`, `int[]`},

		{`int[]`, `float[]`},
		{`int[]`, `mixed[]`},

		{`int|float`, `string`},
		{`int|float`, `\Foo`},
		{`float|int`, `int[]`},
		{`int|float`, `float[]`},
		{`int|float`, `mixed`},

		{`!(int|float)`, `int`},
		{`!(int|float)`, `float`},
		{`!(object|array)`, `object`},
		{`!(object|array)`, `int[]`},
		{`!(object|array)`, `\Foo`},
		{`!(object|array)`, `\Foo[]`},

		{`object|array`, `int|int[]`},
		{`object|array`, `object|float`},
		{`object|array`, `string|float`},
	}

	runTests := func(want bool, tests []testCase) {
		p := phpdoc.NewTypeParser()
		for _, test := range tests {
			val := p.Parse(test.val).Clone()
			dst := p.Parse(test.dst).Clone()

			have := TypeIsCompatible(dst.Expr, val.Expr)
			if have != want {
				t.Errorf("incorrect result: compatible(%s, %s) => %v",
					test.dst, test.val, have)
			}
		}
	}

	runTests(true, matchingTests)
	runTests(false, nonMatchingTests)
}
