package linter

import (
	"fmt"
	"testing"

	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/phpdoc"
)

func TestParseClassPHPDoc(t *testing.T) {
	tests := []struct {
		line     string
		method   string
		property string
		typ      string
	}{
		{
			line:   `@method int foo`,
			method: `foo`,
			typ:    `int`,
		},
		{
			line:   `@method Foo m()`,
			method: `m`,
			typ:    `\Foo`,
		},
		{
			line:   `@method \A\B m2(int $x, ...$rest)`,
			method: `m2`,
			typ:    `\A\B`,
		},
		{
			line:   `@method integer m()`,
			method: `m`,
			typ:    `int`,
		},

		{
			line:     `@property $x int`,
			property: `x`,
			typ:      `int`,
		},
		{
			line:     `@property int $x`,
			property: `x`,
			typ:      `int`,
		},
		{
			line:     `@property int $cost -- item cost in $`,
			property: `cost`,
			typ:      `int`,
		},
		{
			line:     `@property $enabled boolean`,
			property: `enabled`,
			typ:      `bool`,
		},
	}

	st := &meta.ClassParseState{}
	for _, test := range tests {
		doc := fmt.Sprintf(`/** %s */`, test.line)
		result := parseClassPHPDoc(st, doc)

		switch {
		case test.method != "":
			if result.methods.Len() != 1 {
				t.Errorf("parse(`%s`): expected 1 methods, found %d", test.line, result.methods.Len())
				continue
			}
			m, ok := result.methods.Get(test.method)
			if !ok {
				foundInstead := ""
				for _, info := range result.methods.H {
					foundInstead = info.Name
					break
				}
				t.Errorf("parse(`%s`): expected %s method, found %s", test.line, test.method, foundInstead)
				continue
			}
			if m.Typ.String() != test.typ {
				t.Errorf("parse(`%s`): expected %s type, found %s", test.line, test.typ, m.Typ.String())
				continue
			}

		case test.property != "":
			if len(result.properties) != 1 {
				t.Errorf("parse(`%s`): expected 1 properties, found %d", test.line, len(result.properties))
				continue
			}
			m, ok := result.properties[test.property]
			if !ok {
				foundInstead := ""
				for name := range result.properties {
					foundInstead = name
					break
				}
				t.Errorf("parse(`%s`): expected %s property, found %s", test.line, test.property, foundInstead)
				continue
			}
			if m.Typ.String() != test.typ {
				t.Errorf("parse(`%s`): expected %s type, found %s", test.line, test.typ, m.Typ.String())
				continue
			}
		}
	}
}

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
