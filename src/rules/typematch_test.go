package rules

import (
	"testing"

	"github.com/VKCOM/noverify/src/phpdoc"
)

func TestTypeIsCompatible(t *testing.T) {
	tests := []struct {
		dst    string
		val    string
		result bool
	}{
		// Identical types are always compatible.
		{`int`, `int`, true},
		{`string`, `string`, true},
		{`int`, `string`, false},

		// Parens do not change the semantics (but they do affect parsing).
		{`(int)`, `int`, true},
		{`((int))`, `int`, true},
		{`int`, `(int)`, true},
		{`((int))`, `string`, false},

		// "object" special type is compatible with any kind of object.
		{`object`, `object`, true},
		{`object`, `\Foo`, true},
		{`object`, `\Bar`, true},
		{`object`, `string`, false},

		// "array" special type is compatible with any kind of array.
		{`array`, `mixed[]`, true},
		{`array`, `int[]`, true},
		{`array`, `string`, false},

		// Nullable types ?T are compatible with T and null.
		{`?int`, `?int`, true},
		{`?int`, `int`, true},
		{`?int`, `null`, true},
		{`?(int|float)`, `int`, true},
		{`?(int|float)`, `float`, true},
		{`?(int|float)`, `null`, true},
		{`?(int|float)`, `?(int|float)`, true},
		{`?(int|float)`, `?(float|int)`, true},
		{`?(int|float)`, `string`, false},

		// For union types all alternatives are compatible.
		{`int|string`, `int`, true},
		{`string|int`, `int`, true},
		{`int|string`, `string`, true},
		{`string|int`, `string`, true},
		{`string|int`, `float`, false},
		{`\A|\B|\C`, `\A`, true},
		{`\A|\B|\C`, `\B`, true},
		{`\A|\B|\C`, `\C`, true},
		{`\A|\B|\C`, `\D`, false},

		// Union types can be matched by the identical union types.
		// The order of the alternatives doesn't matter.
		{`\A|\B|\C`, `\A|\B|\C`, true},
		{`\A|\C|\B`, `\A|\B|\C`, true},
		{`\C|\B|\A`, `\A|\B|\C`, true},

		// When a single type is asserted against a union, it always fails.
		{`int`, `int|null`, false},
		{`string`, `int|null`, false},

		// When dst and val are unions, we assume val compatible
		// if all of its variants are compatible.
		{`int|float`, `int|float`, true},
		{`int|float`, `float|int`, true},
		{`object|array`, `\Foo|int[]`, true},
		{`object|array`, `int[]|\Foo`, true},
		{`object|array`, `object|float`, false},
		{`object|array`, `float|object`, false},
		{`object|array`, `int|int[]`, false},
		{`int|float`, `string|\A|int`, false},
		{`int|float`, `string|float|\B`, false},
		{`object|array`, `string|float`, false},

		// Type negation inverts the function result.
		{`!int`, `int`, false},
		{`!int`, `string`, true},
		{`!(int|float)`, `int`, false},
		{`!(int|float)`, `float`, false},
		{`!(int|float)`, `int|float`, false},
		{`!(int|float)`, `(float|int)`, false},
		{`!(int|float)`, `string`, true},

		// Double negation works, but no one should ever use that.
		{`!!int`, `int`, true},

		// Mixed val type is assumed to be incompatible with everything.
		// The negation doesn't change that fact.
		{`object`, `mixed`, false},
		{`!object`, `mixed`, false},

		// This is a little compicated case.
		// We resolve it to false for now.
		{`int|null`, `?int`, false},

		// TODO:
		// {`?int`, `int|null`, true},
	}

	p := phpdoc.NewTypeParser()
	for _, test := range tests {
		dstType := p.Parse(test.dst).Clone()
		valType := p.Parse(test.val).Clone()
		have := TypeIsCompatible(dstType.Expr, valType.Expr)
		want := test.result
		if have != want {
			t.Errorf("incorrect result: compatible(%s, %s) => %v",
				test.dst, test.val, have)
		}
	}
}
