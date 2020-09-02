package linter

import (
	"fmt"
	"testing"

	"github.com/VKCOM/noverify/src/meta"
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
	walker := RootWalker{ctx: newRootContext(NewWorkerContext(), st)}
	for _, test := range tests {
		doc := fmt.Sprintf(`/** %s */`, test.line)
		result := walker.parseClassPHPDoc(nil, doc)

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
