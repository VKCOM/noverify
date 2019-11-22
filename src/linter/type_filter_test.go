package linter

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
		{`array`, `\Foo[]`},

		{`object`, `object`},
		{`object`, `\Foo`},
		{`object`, `\Foo\Bar`},

		{`!int`, `string`},
		{`!int`, `mixed`},
		{`!array`, `int`},
		{`!array`, `string`},

		{`int[]`, `int[]`},

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

		// TODO: make union comparison work in these cases.
		// {`a|b|c`, `a|b|c`},
		// {`a|c|b`, `a|b|c`},
		// {`c|b|a`, `a|b|c`},
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
		var p phpdoc.TypeParser
		for _, test := range tests {
			val, err := p.ParseType(test.val)
			if err != nil {
				t.Errorf("parse type `%s`: %v", test.val, err)
				continue
			}
			dst, err := p.ParseType(test.dst)
			if err != nil {
				t.Errorf("parse type `%s`: %v", test.dst, err)
				continue
			}

			have := typeExprIsCompatible(dst, val)
			if have != want {
				t.Errorf("incorrect result: compatible(%s, %s) => %v",
					test.dst, test.val, have)
			}
		}
	}

	runTests(true, matchingTests)
	runTests(false, nonMatchingTests)
}
