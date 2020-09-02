package linter

import (
	"testing"

	"github.com/VKCOM/noverify/src/meta"
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
		{`?int`, `int|null`, true},
		{`?(int|float)`, `int`, true},
		{`?(int|float)`, `float`, true},
		{`?(int|float)`, `null`, true},
		{`?(int|float)`, `string`, false},
		{`?(int|float)`, `?(int|float)`, true},
		{`?(int|float)`, `?(float|int)`, true},

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
		{`a|b|c`, `a|b|c`, true},
		{`a|c|b`, `a|b|c`, true},
		{`c|b|a`, `a|b|c`, true},

		// This is a little compicated case.
		// We resolve it to false for now.
		{`int|null`, `?int`, false},

		// When val is union we treat it as a list of types to be tried on.
		{`int`, `int|null`, true},
		{`string`, `int|null`, false},

		// When dst and val are unions, we assume val compatible
		// if any of its variants are compatible.
		{`int|float`, `int|float`, true},
		{`int|float`, `float|int`, true},
		{`int|float`, `string|\A|int`, true},
		{`int|float`, `string|float|\B`, true},
		{`object|array`, `int|int[]`, true},
		{`object|array`, `object|float`, true},
		{`object|array`, `float|object`, true},
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
	}

	p := phpdoc.NewTypeParser()
	for _, test := range tests {
		dstType := p.Parse(test.dst).Clone()
		valType := p.Parse(test.val).Clone()
		have := typeIsCompatible(dstType.Expr, valType.Expr)
		want := test.result
		if have != want {
			t.Errorf("incorrect result: compatible(%s, %s) => %v",
				test.dst, test.val, have)
		}
	}
}

func BenchmarkParseTypes(b *testing.B) {
	st := &meta.ClassParseState{}
	ctx := newRootContext(NewWorkerContext(), st)
	typeString := `?x|array<int>|T[]`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsedType := ctx.phpdocTypeParser.Parse(typeString)
		types, _ := typesFromPHPDoc(&ctx, parsedType)
		_ = newTypesMap(&ctx, types)
	}
}
